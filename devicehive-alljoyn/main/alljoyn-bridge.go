package main

/*
#cgo CFLAGS: -I../ajtcl/inc -I../ajtcl/target/linux
#cgo LDFLAGS: -L../ajtcl -lajtcl
#include <stdio.h>
#include <aj_debug.h>
#include <aj_guid.h>
#include <aj_creds.h>
#include <aj_peer.h>
#include <aj_link_timeout.h>
#include "alljoyn.h"

uint32_t Get_AJ_Message_msgId();
uint32_t Get_AJ_Message_bodyLen();
const char * Get_AJ_Message_signature();
const char * Get_AJ_Message_objPath();
const char * Get_AJ_Message_iface();
const char * Get_AJ_Message_member();
const char * Get_AJ_Message_destination();
void * Get_AJ_ReplyMessage();
void * Get_AJ_Message();
void * Get_AJ_BusAttachment();
void * Allocate_AJ_Object_Array(uint32_t array_size);
void * Create_AJ_Object(uint32_t index, void * array, char* path, AJ_InterfaceDescription* interfaces, uint8_t flags, void* context);
AJ_Status MyAboutPropGetter_cgo(AJ_Message* reply, const char* language);
void * Get_Session_Opts();
void * Get_Arg();
AJ_Status AJ_MarshalArgs_cgo(AJ_Message* msg, char * a, char * b, char * c, char * d);
int UnmarshalPort();
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"github.com/devicehive/IoT-framework/devicehive-alljoyn/ajmarshal"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

const PORT = 42
const AJ_CP_PORT = 1000


/* Gloabal variables to preserve from garbage collection to pass safely to cgo */
var interfaces []C.AJ_InterfaceDescription
var myPropGetterFunction PropGetterFunction
var myMessenger *AllJoynMessenger

type PropGetterFunction func(reply *C.AJ_Message, language *C.char) C.AJ_Status

type AllJoynBindingInfo struct {
	allJoynService string
	allJoynPath    string
	dbusPath       string
	introspectData *introspect.Node
}

type AllJoynBridge struct {
	bus                *dbus.Conn
	services           map[string][]*AllJoynBindingInfo
}

type AllJoynMessenger struct {
	dbusService string
	bus         *dbus.Conn
	binding     []*AllJoynBindingInfo
}

func NewAllJoynMessenger(dbusService string, bus *dbus.Conn, binding []*AllJoynBindingInfo) *AllJoynMessenger {
	return &AllJoynMessenger{dbusService, bus, binding}
}

func NewAllJoynBridge(bus *dbus.Conn) *AllJoynBridge {
	bridge := new(AllJoynBridge)
	bridge.bus = bus
	bridge.services = make(map[string][]*AllJoynBindingInfo)

	return bridge
}

func ParseArgumentOrProperty(name, access, _type string) string {
	s := name
	if access == "in" || access == "write" {
		s = s + "<"
	} else if access == "out" || access == "read" {
		s = s + ">"
	} else if access == "" {
		s = s + ">" // For signals
	} else {
		s = s + "="
	}
	s = s + _type
	return s
}

func ParseArguments(args []introspect.Arg) string {
	argString := ""
	for _, arg := range args {
		argString = argString + " " + ParseArgumentOrProperty(arg.Name, arg.Direction, arg.Type)
	}
	return argString
}

func ParseAllJoynInterfaces(interfaces []introspect.Interface) []C.AJ_InterfaceDescription {
	res := make([]C.AJ_InterfaceDescription, 0)

	for _, iface := range interfaces {
		desc := make([]*C.char, 0)
		desc = append(desc, C.CString(iface.Name))

		for _, method := range iface.Methods {
			methogString := "?" + method.Name
			argString := ParseArguments(method.Args)
			log.Print(methogString + argString)
			desc = append(desc, C.CString(methogString+argString))
		}

		for _, signal := range iface.Signals {
			signalString := "!" + signal.Name
			argString := ParseArguments(signal.Args)
			log.Print(signalString + argString)
			desc = append(desc, C.CString(signalString+argString))
		}

		for _, prop := range iface.Properties {
			propString := "@" + ParseArgumentOrProperty(prop.Name, prop.Access, prop.Type)
			log.Print(propString)
			desc = append(desc, C.CString(propString))
		}

		desc = append(desc, nil)
		log.Print(desc)
		res = append(res, (C.AJ_InterfaceDescription)(&desc[0]))
	}
	res = append(res, nil)
	return res
}

func GetAllJoynObjects(services []*AllJoynBindingInfo) unsafe.Pointer {
	array := C.Allocate_AJ_Object_Array(C.uint32_t(len(services) + 1))
	for i, service := range services {
		interfaces = ParseAllJoynInterfaces(service.introspectData.Interfaces)
		C.Create_AJ_Object(C.uint32_t(i), array, C.CString(service.introspectData.Name), &interfaces[0], C.AJ_OBJ_FLAG_ANNOUNCED, unsafe.Pointer(nil))
		C.Create_AJ_Object(C.uint32_t(i+1), array, nil, nil, 0, nil)
		log.Printf("*****Alljoyn Objects %s %s", service.introspectData.Name, &interfaces[0])
	}
	return array
}

//export MyAboutPropGetter
func MyAboutPropGetter(reply *C.AJ_Message, language *C.char) C.AJ_Status {
	return myMessenger.MyAboutPropGetter_member(reply, language)
}

func (m *AllJoynMessenger) MyAboutPropGetter_member(reply *C.AJ_Message, language *C.char) C.AJ_Status {
	log.Printf("MyAboutPropGetter_member(): %+v", m)
	aboutInterfacePath := m.binding[0].dbusPath

	log.Printf("About interface path: %s", aboutInterfacePath)
	err := m.callRemoteMethod(reply, aboutInterfacePath, "org.alljoyn.About.GetAboutData", []interface{}{C.GoString(language)})
	if err != nil {
		log.Printf("Error calling org.alljoyn.About for [%+v]: %s", m.binding, err)
		return C.AJ_ERR_NO_MATCH
	}

	log.Printf("About message signature, offset: %s, %d", C.GoString(reply.signature), reply.sigOffset)
	return C.AJ_OK
}

func (m *AllJoynMessenger) callRemoteMethod(message *C.AJ_Message, path, member string, arguments []interface{}) (err error) {
	remote := m.bus.Object(m.dbusService, dbus.ObjectPath(path))
	log.Printf("%s Argument[0] %+v", member, reflect.ValueOf(arguments[0]).Type())
	res := remote.Call(member, 0, arguments...)

	if res.Err != nil {
		log.Printf("Error calling dbus method (%s): %s", member, res.Err)
		return res.Err
	}

	//pad := 4 - (int)(message.bodyBytes)%4
	//C.WriteBytes((*C.AJ_Message)(message), nil, (C.size_t)(0), (C.size_t)(pad))

	buf := new(bytes.Buffer)
	buf.Write(make([]byte, (int)(message.bodyBytes)))
	enc := devicehivealljoyn.NewEncoderAtOffset(buf, (int)(message.bodyBytes), binary.LittleEndian)
	pad, err := enc.Encode(res.Body...)
	log.Printf("Padding of the encoded buffer: %d", pad)
	//log.Printf("Got reply: %+v", res.Body)
	if err != nil {
		log.Printf("Error encoding result: %s", err)
		return err
	}
	/*
		C.AJ_MarshalContainer((*C.AJ_Message)(message), (*C.AJ_Arg)(C.Get_Arg()), C.AJ_ARG_ARRAY)

		C.AJ_MarshalArgs_cgo((*C.AJ_Message)(message), C.CString("{sv}"), C.CString("DeviceName"), C.CString("s"), C.CString("Golang-device"))

		C.AJ_MarshalCloseContainer((*C.AJ_Message)(message), (*C.AJ_Arg)(C.Get_Arg()))
	*/
	//log.Printf("Encoded reply, len: %+v, %d", buf.Bytes(), buf.Len())
	log.Printf("Length before: %d", message.bodyBytes)

	newBuf := buf.Bytes()[(int)(message.bodyBytes)+pad:]
	//log.Printf("Buffer to write into AllJoyn: %+v, %d", newBuf, len(newBuf))
	//	hdr := message.hdr
	//	if hdr.flags&C.uint8_t(C.AJ_FLAG_ENCRYPTED) == 0 {
	//		hdr = nil
	//	} else {
	//		message.hdr.flags &= ^C.uint8_t(C.AJ_FLAG_ENCRYPTED)
	//	}
	C.AJ_DeliverMsgPartial((*C.AJ_Message)(message), C.uint32_t(len(newBuf)))
	C.AJ_MarshalRaw((*C.AJ_Message)(message), unsafe.Pointer(&newBuf[0]), C.size_t(len(newBuf)))
	//log.Printf("New buff reply, len: %+v, %d", newBuf, len(newBuf))
	//	if hdr != nil {
	//		message.hdr = hdr
	//		message.hdr.flags &= ^C.uint8_t(C.AJ_FLAG_ENCRYPTED)
	//	}
	return nil
}

func safeString(p * C.char ) string{	
	if p == nil {
		return ""
	} else {
		return C.GoString(p)
	}
}

func dumpMessage(prefix string){

	msgId := uint32(C.Get_AJ_Message_msgId())
	objPath := "" //safeString(C.Get_AJ_Message_objPath())
	member := safeString(C.Get_AJ_Message_member())
	destination := safeString(C.Get_AJ_Message_destination())
	iface := safeString(C.Get_AJ_Message_iface())

	log.Printf("%s \r\n\tmsgId: %d \r\n\tobjPath: %s \r\n\tmember: %s \r\n\tiface: %s \r\n\tdestination: %s", prefix, msgId, objPath, member, iface, destination)

}

func (m *AllJoynMessenger) forwardAllJoynMessage(msgId uint32) (err error) {
	log.Printf("****forwardAllJoynMessage****")
	msg := C.Get_AJ_Message()
	reply := C.Get_AJ_ReplyMessage()

	var data uintptr
	var actual C.size_t

	b := make([]byte, 0)
	signature := C.GoString(C.Get_AJ_Message_signature())
	bodyLen := C.Get_AJ_Message_bodyLen()
	for i := 0; i < int(bodyLen); i++ {
		status := C.AJ_UnmarshalRaw((*C.AJ_Message)(msg), (*unsafe.Pointer)(unsafe.Pointer(&data)), C.size_t(1), (*C.size_t)(unsafe.Pointer(&actual)))
		if status == C.AJ_OK {
			b = append(b, C.GoBytes(unsafe.Pointer(data), C.int(actual))...)
			log.Printf("Reading RAW message, status = %d, actual = %d", status, actual)
		} else {
			log.Printf("Error while reading message body, status = %d", status)
			break
		}
	}
	s, err := dbus.ParseSignature(signature)

	if err != nil {
		log.Printf("Error parsing signature: %s", err)
		return err
	}

	d := dbus.NewDecoder(bytes.NewReader(b), binary.LittleEndian)
	res, err := d.Decode(s)

	if err != nil {
		log.Printf("Error decoding message [%+v] : %s", b, err)
		return err
	}

	//log.Printf("Received application alljoyn message, signature: %s, bytes: %+v, decoded: %+v", signature, b, res)

	objPath := C.GoString(C.Get_AJ_Message_objPath())
	member := C.GoString(C.Get_AJ_Message_member())
	destination := C.GoString(C.Get_AJ_Message_destination())
	iface := C.GoString(C.Get_AJ_Message_iface())

	log.Printf("****Message [objPath, member, iface, destination]: %s, %s, %s, %s", objPath, member, iface, destination)

	for _, service := range m.binding {
		if service.allJoynPath == objPath {
			log.Print("Found matching dbus service: %+v", service)
			C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
			m.callRemoteMethod((*C.AJ_Message)(reply), service.dbusPath, iface+"."+member, res)
			C.AJ_DeliverMsg((*C.AJ_Message)(reply))
			break
		}
	}
	return nil
}

func (a *AllJoynBridge) StartAllJoyn(dbusService string) *dbus.Error {
	services := a.services[dbusService]
	objects := GetAllJoynObjects(services)

	myMessenger = NewAllJoynMessenger(dbusService, a.bus, services)

	log.Printf("StartAllJoyn: %+v", myMessenger)

	C.AJ_Initialize()
	C.AJ_RegisterObjects((*C.AJ_Object)(objects), nil)
	C.AJ_AboutRegisterPropStoreGetter((C.AJ_AboutPropGetter)(unsafe.Pointer(C.MyAboutPropGetter_cgo)))
	C.AJ_SetMinProtoVersion(10)

	C.AJ_PrintXML((*C.AJ_Object)(objects))
	connected := false
	var status C.AJ_Status = C.AJ_OK
	busAttachment := C.Get_AJ_BusAttachment()
	msg := C.Get_AJ_Message()
	C.AJ_ClearAuthContext()

	log.Printf("CreateAJ_BusAttachment(): %+v", busAttachment)

	go func() {
		for {
			if !connected {
				status = C.AJ_StartService((*C.AJ_BusAttachment)(busAttachment),
					C.CString("org.alljoyn.BusNode"),
					60*1000, // TODO: Move connection timeout to config
					C.FALSE,
					C.uint16_t(PORT), // TODO: Move port to config
					C.CString(services[0].allJoynService),
					C.AJ_NAME_REQ_DO_NOT_QUEUE,
					(*C.AJ_SessionOpts)(C.Get_Session_Opts()))

				if status != C.AJ_OK {
					continue
				}

				log.Printf("StartService returned %d, %+v", status, busAttachment)

				connected = true

				status = C.AJ_BusBindSessionPort((*C.AJ_BusAttachment)(busAttachment), 
					AJ_CP_PORT, (*C.AJ_SessionOpts)(C.Get_Session_Opts()), 0);

				if (status != C.AJ_OK) {
					log.Printf(("Failed to send bind session port message"));
				}

			}


			status = C.AJ_UnmarshalMsg((*C.AJ_BusAttachment)(busAttachment), (*C.AJ_Message)(msg),
				5*1000) // TODO: Move unmarshal timeout to config
			log.Printf("AJ_UnmarshalMsg: %+v", status)

			if C.AJ_ERR_TIMEOUT == status {
				continue
			}

			if C.AJ_OK == status {

				msgId := C.Get_AJ_Message_msgId()
				log.Printf("****Got a message, ID: 0x%X", msgId)
				dumpMessage("****Message Detais: ")

				switch {
				case msgId == C.AJ_METHOD_ACCEPT_SESSION:
					{
						port := int(C.UnmarshalPort())
						if port == PORT || port == AJ_CP_PORT {
							status = C.AJ_BusReplyAcceptSession((*C.AJ_Message)(msg), C.TRUE)
							log.Printf("ACCEPT_SESSION: %+v port %d", msgId, port)
						} else {
							status = C.AJ_BusReplyAcceptSession((*C.AJ_Message)(msg), C.FALSE)
							log.Printf("REJECT_SESSION: %+v port", msgId, port)
						}
					}

				case msgId == C.AJ_SIGNAL_SESSION_LOST_WITH_REASON:
					{
						// uint32_t id, reason;
						// AJ_UnmarshalArgs(&msg, "uu", &id, &reason);
						// AJ_AlwaysPrintf(("Session lost. ID = %u, reason = %u", id, reason));
						log.Printf("Session lost: %+v", msgId)
					}
				case uint32(msgId) == 0x1010003: // Config.GetConfigurations
					// our forwardAllJoyn doesn't support encrypted messages which config service is,
					// so we handle it here manually
					{
						//C.AJ_UnmarshalArgs(msg, "s", &language);
						reply := C.Get_AJ_ReplyMessage()
						C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
						C.AJ_MarshalContainer((*C.AJ_Message)(reply), (*C.AJ_Arg)(C.Get_Arg()), C.AJ_ARG_ARRAY)
						C.AJ_MarshalArgs_cgo((*C.AJ_Message)(reply), C.CString("{sv}"), C.CString("DeviceName"), C.CString("s"), C.CString("DeviceHiveVB"))
						C.AJ_MarshalArgs_cgo((*C.AJ_Message)(reply), C.CString("{sv}"), C.CString("DefaultLanguage"), C.CString("s"), C.CString("en"))
						C.AJ_MarshalCloseContainer((*C.AJ_Message)(reply), (*C.AJ_Arg)(C.Get_Arg()))
						C.AJ_DeliverMsg((*C.AJ_Message)(reply))
					}
				case (uint32(msgId) & 0x01000000) != 0:
					{
						myMessenger.forwardAllJoynMessage(uint32(msgId))
					}
				default:

					if (uint32(msgId) & 0xFFFF0000) == 0x00050000 {
						if uint32(msgId) == 0x00050102 {
							log.Printf("Passing About.GetObjectDescription %+v to AllJoyn", msgId)
							status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
						} else if uint32(msgId) == 0x00050101 {
							log.Printf("Passing About.GetAboutData %+v to AllJoyn", msgId)
							status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
						} else if uint32(msgId) == 0x00050000 {
							log.Printf("Passing Properties.Get %+v to AllJoyn", msgId)
							status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
						} else {
							myMessenger.forwardAllJoynMessage(uint32(msgId))
						}
					} else {
						/* Pass to the built-in handlers. */
						log.Printf("Passing msgId %+v to AllJoyn", uint32(msgId))
						status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
						log.Printf("AllJoyn returned %d", status)
					}
				}
				C.AJ_NotifyLinkActive()
			}

			/* Messages MUST be discarded to free resources. */
			C.AJ_CloseMsg((*C.AJ_Message)(msg))

			if status == C.AJ_OK {
				log.Print("***C.AJ_AboutAnnounce***")
				C.AJ_AboutAnnounce((*C.AJ_BusAttachment)(busAttachment))
			}

			if status == C.AJ_ERR_READ {
				C.AJ_Disconnect((*C.AJ_BusAttachment)(busAttachment))
				log.Print("AllJoyn disconnected, retrying")
				connected = false
				C.AJ_Sleep(1000 * 2) // TODO: Move sleep time to const
			}
		}
	}()
	return nil
}

func (a *AllJoynBridge) addService(service string, info *AllJoynBindingInfo) {
	services, ok := a.services[service]
	if ok {
		a.services[service] = append(services, info)
	} else {
		a.services[service] = []*AllJoynBindingInfo{info}
	}
}

func (a *AllJoynBridge) AddService(dbusPath, dbusService, allJoynPath, allJoynService, introspectXml string) *dbus.Error {	
	var xmldata string 
	var node introspect.Node

	if introspectXml != "" {
		// introspect data was provided with method call
		xmldata = introspectXml		
	} else {
		// get introspecction data from dbus
		var o = a.bus.Object(dbusService, dbus.ObjectPath(dbusPath))		
		err := o.Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&xmldata)		
		if err != nil {
			log.Printf("Error getting introspect from [%s, %s]: %s", dbusService, dbusPath, err)
			return nil
		}
	}

	err := xml.NewDecoder(strings.NewReader(xmldata)).Decode(&node)

	if err != nil {
		log.Printf("Error decoding introspect data for [%s, %s]: %s", dbusService, dbusPath, err)
		return nil
	}

	a.addService(dbusService, &AllJoynBindingInfo{allJoynService, allJoynPath, dbusPath, &node})

	log.Printf("Received introspect: %+v", &node)

	return nil
}

func main() {
	bus, err := dbus.SystemBus()

	if err != nil {
		log.Fatal(err)
	}

	res, err := bus.RequestName("com.devicehive.alljoyn.bridge",
		dbus.NameFlagDoNotQueue)

	if err != nil {
		log.Fatalf("Failed to request dbus name: %s", err)
	}

	if res != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalf("Failed to request dbus name: %+v", res)
	}

	allJoynBridge := NewAllJoynBridge(bus)

	bus.Export(allJoynBridge, "/com/devicehive/alljoyn/bridge", "com.devicehive.alljoyn.bridge")

	n := &introspect.Node{
		Interfaces: []introspect.Interface{
			{
				Name:    "com.devicehive.alljoyn.bridge",
				Methods: introspect.Methods(allJoynBridge),
				Signals: []introspect.Signal{},
			},
		},
	}

	root := &introspect.Node{
		Children: []introspect.Node{
			{
				Name:    "com/devicehive/alljoyn/bridge",
			},
		},
	}

	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/alljoyn/bridge", "org.freedesktop.DBus.Introspectable")
	bus.Export(introspect.NewIntrospectable(root), "/", "org.freedesktop.DBus.Introspectable") // workaroud for dbus issue #14

	log.Printf("Bridge is Running.")

	select {}
}
