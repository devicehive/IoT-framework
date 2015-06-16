package main

/*
#cgo CFLAGS: -I../ajtcl/inc -I../ajtcl/target/linux
#cgo LDFLAGS: -L../ajtcl -lajtcl
#include <stdio.h>
#include <aj_debug.h>
#include <aj_guid.h>
#include <aj_creds.h>
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
void * Get_Arg();
AJ_Status AJ_MarshalArgs_cgo(AJ_Message* msg, char * a, char * b, char * c, char * d);
AJ_MsgHeader * BackupMsgHeader(AJ_MsgHeader * src);
*/
import "C"
import (
	"github.com/devicehive/IoT-framework/devicehive-alljoyn/ajmarshal"
	"bytes"
	"encoding/binary"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
	"unsafe"
)

/* Gloabal variables to preserve from garbage collection to pass safely to cgo */
var interfaces []C.AJ_InterfaceDescription
var myPropGetterFunction PropGetterFunction
var myMessenger *AllJoynMessenger

type IntrospectProvider func(dbusService, dbusPath string) (node *introspect.Node, err error)
type PropGetterFunction func(reply *C.AJ_Message, language *C.char) C.AJ_Status

type AllJoynBindingInfo struct {
	allJoynService string
	allJoynPath    string
	dbusPath       string
	introspectData *introspect.Node
}

type AllJoynBridge struct {
	bus                *dbus.Conn
	introspectProvider IntrospectProvider
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

func NewAllJoynBridge(bus *dbus.Conn, introspectProvider IntrospectProvider) *AllJoynBridge {
	bridge := new(AllJoynBridge)
	bridge.bus = bus
	bridge.services = make(map[string][]*AllJoynBindingInfo)
	bridge.introspectProvider = introspectProvider

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
	}

	return array
}

//export MyAboutPropGetter
func MyAboutPropGetter(reply *C.AJ_Message, language *C.char) C.AJ_Status {
	return myMessenger.MyAboutPropGetter_member(reply, language)
}

func (m *AllJoynMessenger) MyAboutPropGetter_member(reply *C.AJ_Message, language *C.char) C.AJ_Status {
	// Find about path for the bindibg
	log.Printf("MyAboutPropGetter_member(): %+v", m)
	// var aboutInterface *introspect.Interface
	aboutInterfacePath := m.binding[0].dbusPath
	// for _, b := range m.binding {
	// 	for _, iface := range b.introspectData.Interfaces {
	// 		if iface.Name == "org.AllJoyn.About" {
	// 			aboutInterfacePath = b.dbusPath
	// 			aboutInterface = &iface
	// 			break
	// 		}
	// 	}
	// }

	// if aboutInterface == nil {
	// 	log.Printf("No org.AllJoyn.About interface found for %+v", m.binding)
	// 	return C.AJ_ERR_NO_MATCH
	// }

	log.Printf("About interface path: %s", aboutInterfacePath)
	err := m.callRemoteMethod(reply, aboutInterfacePath, "org.alljoyn.About.GetAboutData", C.GoString(language))
	if err != nil {
		log.Printf("Error calling org.alljoyn.About for [%+v]: %s", m.binding, err)
		return C.AJ_ERR_NO_MATCH
	}

	log.Printf("About message signature, offset: %s, %d", C.GoString(reply.signature), reply.sigOffset)
	return C.AJ_OK
}

func (m *AllJoynMessenger) callRemoteMethod(message *C.AJ_Message, path, member string, arguments ...interface{}) (err error) {
	remote := m.bus.Object(m.dbusService, dbus.ObjectPath(path))
	res := remote.Call(member, 0, arguments...)

	if res.Err != nil {
		log.Printf("Error calling dbus method (%s): %s", member, res.Err)
		return res.Err
	}

	buf := new(bytes.Buffer)
	enc := devicehivealljoyn.NewEncoder(buf, binary.LittleEndian)
	err = enc.Encode(res.Body...)
	log.Printf("Got reply: %+v", res.Body)
	if err != nil {
		log.Printf("Error encoding result: %s", err)
		return err
	}

	C.AJ_MarshalContainer((*C.AJ_Message)(message), (*C.AJ_Arg)(C.Get_Arg()), C.AJ_ARG_ARRAY)

	C.AJ_MarshalArgs_cgo((*C.AJ_Message)(message), C.CString("{sv}"), C.CString("DeviceName"), C.CString("s"), C.CString("Golang-device"))

	C.AJ_MarshalCloseContainer((*C.AJ_Message)(message), (*C.AJ_Arg)(C.Get_Arg()))

	log.Printf("Encoded reply, len: %+v, %d", buf.Bytes(), buf.Len())
	log.Printf("Length before: %d", message.bodyBytes)
	// message.hdr.bodyLen = (C.uint32_t)(message.bodyBytes + C.uint16_t(buf.Len()))
	// message.bodyBytes = (C.uint16_t)(message.hdr.bodyLen) + (C.uint16_t)(buf.Len())
	// oldHeader := C.BackupMsgHeader(message.hdr)
	// message.hdr = nil
	//C.AJ_DeliverMsgPartial((*C.AJ_Message)(message), C.uint32_t(buf.Len()))
	// log.Printf("Length before: %d", message.hdr.bodyLen)
	// log.Printf("BodyBytes, hdr.BodyLen: %d, %d", message.bodyBytes, message.hdr.bodyLen)
	// message.bodyBytes += C.uint16_t(buf.Len())
	// message.hdr.bodyLen = C.uint32_t(message.bodyBytes)
	// message.hdr = nil
	//C.AJ_MarshalRaw((*C.AJ_Message)(message), unsafe.Pointer(&buf.Bytes()[0]), C.size_t(buf.Len()))
	// message.hdr = oldHeader
	return nil
}

func (m *AllJoynMessenger) forwardAllJoynMessage(msgId uint32) (err error) {
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

	d := devicehivealljoyn.NewDecoder(bytes.NewReader(b), binary.LittleEndian)
	res, err := d.Decode(s)

	if err != nil {
		log.Printf("Error decoding message [%+v] : %s", b, err)
		return err
	}

	log.Printf("Received application alljoyn message, signature: %s, bytes: %+v, decoded: %+v", signature, b, res)

	objPath := C.GoString(C.Get_AJ_Message_objPath())
	member := C.GoString(C.Get_AJ_Message_member())
	destination := C.GoString(C.Get_AJ_Message_destination())
	iface := C.GoString(C.Get_AJ_Message_iface())

	log.Printf("Message [objPath, member, iface, destination]: %s, %s, %s, %s", objPath, member, iface, destination)

	for _, service := range m.binding {
		if service.allJoynPath == objPath {
			log.Print("Found matching dbus service: %+v", service)
			C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
			m.callRemoteMethod((*C.AJ_Message)(reply), service.dbusPath, iface+"."+member, res)
			// remote := m.bus.Object(dbusService, dbus.ObjectPath(service.dbusPath))
			// res := remote.Call(iface+"."+member, 0, res...)

			// if res.Err != nil {
			// 	log.Printf("Error calling dbus method (%s): %s", iface+"."+member, res.Err)
			// 	return res.Err
			// }

			// C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
			// buf := new(bytes.Buffer)
			// enc := newEncoder(buf, binary.LittleEndian)
			// err = enc.Encode(res.Body...)
			// if err != nil {
			// 	log.Printf("Error encoding result: %s", err)
			// 	break
			// }
			// log.Printf("Encoded reply: %+v", buf.Bytes())
			// C.AJ_DeliverMsgPartial((*C.AJ_Message)(reply), C.uint32_t(buf.Len()))
			// C.AJ_MarshalRaw((*C.AJ_Message)(reply), unsafe.Pointer(&buf.Bytes()[0]), C.size_t(buf.Len()))
			C.AJ_DeliverMsg((*C.AJ_Message)(reply))
			break
		}
	}
	return nil
}

func (a *AllJoynBridge) StartAllJoyn(dbusService string) *dbus.Error {
	service := a.services[dbusService]
	objects := GetAllJoynObjects(a.services[dbusService])

	myMessenger = NewAllJoynMessenger(dbusService, a.bus, a.services[dbusService])

	log.Printf("StartAllJoyn: %+v", myMessenger)

	C.AJ_Initialize()
	C.AJ_RegisterObjects((*C.AJ_Object)(objects), nil)
	C.AJ_AboutRegisterPropStoreGetter((C.AJ_AboutPropGetter)(unsafe.Pointer(C.MyAboutPropGetter_cgo)))

	C.AJ_PrintXML((*C.AJ_Object)(objects))
	connected := false
	var status C.AJ_Status = C.AJ_OK
	busAttachment := C.Get_AJ_BusAttachment()
	msg := C.Get_AJ_Message()

	log.Printf("CreateAJ_BusAttachment(): %+v", busAttachment)

	for {
		if !connected {
			status = C.AJ_StartService((*C.AJ_BusAttachment)(busAttachment),
				nil,
				60*1000, // TODO: Move connection timeout to config
				C.FALSE,
				25, // TODO: Move port to config
				C.CString(service[0].allJoynService),
				C.AJ_NAME_REQ_DO_NOT_QUEUE,
				nil)

			if status != C.AJ_OK {
				continue
			}

			log.Printf("StartService returned %d, %+v", status, busAttachment)

			connected = true
		}

		status = C.AJ_UnmarshalMsg((*C.AJ_BusAttachment)(busAttachment), (*C.AJ_Message)(msg),
			5*1000) // TODO: Move unmarshal timeout to config
		log.Printf("AJ_UnmarshalMsg: %+v", status)

		if C.AJ_ERR_TIMEOUT == status {
			continue
		}

		if C.AJ_OK == status {

			msgId := C.Get_AJ_Message_msgId()
			log.Printf("Received message: %+v", msgId)

			switch {
			case msgId == C.AJ_METHOD_ACCEPT_SESSION:
				{
					// uint16_t port;
					// char* joiner;
					// uint32_t sessionId;

					// AJ_UnmarshalArgs(&msg, "qus", &port, &sessionId, &joiner);
					status = C.AJ_BusReplyAcceptSession((*C.AJ_Message)(msg), C.TRUE)
					log.Printf("ACCEPT_SESSION: %+v", msgId)
				}

			case msgId == C.AJ_SIGNAL_SESSION_LOST_WITH_REASON:
				{
					// uint32_t id, reason;
					// AJ_UnmarshalArgs(&msg, "uu", &id, &reason);
					// AJ_AlwaysPrintf(("Session lost. ID = %u, reason = %u", id, reason));
					log.Printf("Session lost: %+v", msgId)
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
					} else {
						myMessenger.forwardAllJoynMessage(uint32(msgId))
					}
				} else {
					/* Pass to the built-in handlers. */
					log.Printf("Passing msgId %+v to AllJoyn", msgId)
					status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
				}
			}
		}

		/* Messages MUST be discarded to free resources. */
		C.AJ_CloseMsg((*C.AJ_Message)(msg))

		if status == C.AJ_ERR_READ {
			C.AJ_Disconnect((*C.AJ_BusAttachment)(busAttachment))
			log.Print("AllJoyn disconnected, retrying")
			connected = false
			C.AJ_Sleep(1000 * 2) // TODO: Move sleep time to const
		}
	}
	// }()
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

func (a *AllJoynBridge) AddService(dbusPath, dbusService, allJoynPath, allJoynService string) *dbus.Error {
	node, err := a.introspectProvider(dbusService, dbusPath)

	if err != nil {
		log.Printf("Error getting introspect from [%s, %s]: %s", dbusService, dbusPath, err)
	}

	a.addService(dbusService, &AllJoynBindingInfo{allJoynService, allJoynPath, dbusPath, node})

	log.Printf("Received introspect: %+v", node)

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

	allJoynBridge := NewAllJoynBridge(bus, func(dbusService, dbusPath string) (*introspect.Node, error) {
		return introspect.Call(bus.Object(dbusService, dbus.ObjectPath(dbusPath)))
	})

	bus.Export(allJoynBridge, "/com/devicehive/alljoyn/bridge", "com.devicehive.alljoyn.bridge")
	select {}
}
