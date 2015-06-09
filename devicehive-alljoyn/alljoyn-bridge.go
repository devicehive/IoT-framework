package main

// #cgo CFLAGS: -Iajtcl/inc -Iajtcl/target/linux
// #cgo LDFLAGS: -Lajtcl -lajtcl
// #include <stdio.h>
// #include <aj_debug.h>
// #include <aj_guid.h>
// #include <aj_creds.h>
// #include "alljoyn.h"
//
// AJ_BusAttachment c_bus;
// AJ_Message c_message;
// AJ_Message c_reply;
//
// uint32_t Get_AJ_Message_msgId() {
//   return c_message.msgId;
// }
//
// uint32_t Get_AJ_Message_bodyLen() {
//   return c_message.hdr->bodyLen;
// }
//
// const char * Get_AJ_Message_signature() {
//    return c_message.signature;
// }
//
// const char * Get_AJ_Message_objPath() {
//    return c_message.objPath;
// }
//
// const char * Get_AJ_Message_iface() {
//    return c_message.iface;
// }
//
// const char * Get_AJ_Message_member() {
//    return c_message.member;
// }
//
// const char * Get_AJ_Message_destination() {
//    return c_message.destination;
// }
//
// void * Get_AJ_ReplyMessage() {
// 	return &c_reply;
// }
//
// void * Get_AJ_Message() {
//   return &c_message;
// }
// void * Get_AJ_BusAttachment() {
//   return &c_bus;
// }
//
// void * Allocate_AJ_Object_Array(uint32_t array_size) {
//   return AJ_Malloc(sizeof(AJ_Object)*array_size);
// }
//
// void * Create_AJ_Object(uint32_t index, void * array, char* path, AJ_InterfaceDescription* interfaces, uint8_t flags, void* context) {
//	 AJ_Object * obj = array + index * sizeof(AJ_Object);
//   obj->path = path;
//	 obj->interfaces = interfaces;
//   obj->flags = flags;
//   obj->context = context;
//   return obj;
// }
//
//
import "C"
import (
	"bytes"
	"encoding/binary"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
	"unsafe"
)

type IntrospectProvider func(dbusService, dbusPath string) (node *introspect.Node, err error)

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

var interfaces []C.AJ_InterfaceDescription

func GetAllJoynObjects(services []*AllJoynBindingInfo) unsafe.Pointer {
	array := C.Allocate_AJ_Object_Array(C.uint32_t(len(services) + 1))

	for i, service := range services {
		interfaces = ParseAllJoynInterfaces(service.introspectData.Interfaces)
		C.Create_AJ_Object(C.uint32_t(i), array, C.CString(service.introspectData.Name), &interfaces[0], C.uint8_t(0), unsafe.Pointer(nil))
		C.Create_AJ_Object(C.uint32_t(i+1), array, nil, nil, 0, nil)
	}

	return array
}

func (a *AllJoynBridge) StartAllJoyn(dbusService string) *dbus.Error {
	objects := GetAllJoynObjects(a.services[dbusService])
	// go func() {
	C.AJ_Initialize()
	C.AJ_RegisterObjects((*C.AJ_Object)(objects), nil)
	C.AJ_PrintXML((*C.AJ_Object)(objects))
	connected := false
	var status C.AJ_Status = C.AJ_OK
	busAttachment := C.Get_AJ_BusAttachment()
	msg := C.Get_AJ_Message()
	reply := C.Get_AJ_ReplyMessage()

	var data uintptr
	var actual C.size_t

	log.Printf("CreateAJ_BusAttachment(): %+v", busAttachment)

	for {
		if !connected {
			status = C.AJ_StartService((*C.AJ_BusAttachment)(busAttachment),
				nil,
				60*1000, // TODO: Move connection timeout to config
				C.FALSE,
				25, // TODO: Move port to config
				C.CString(dbusService),
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
					b := make([]byte, 0)
					signature := C.GoString(C.Get_AJ_Message_signature())
					bodyLen := C.Get_AJ_Message_bodyLen()
					for i := 0; i < int(bodyLen); i++ {
						status = C.AJ_UnmarshalRaw((*C.AJ_Message)(msg), (*unsafe.Pointer)(unsafe.Pointer(&data)), C.size_t(1), (*C.size_t)(unsafe.Pointer(&actual)))
						if status == C.AJ_OK {
							b = append(b, C.GoBytes(unsafe.Pointer(data), C.int(actual))...)
							log.Printf("Reading RAW message, status = %d, actual = %d", status, actual)
						} else {
							log.Printf("Error while reading message body, status = %d", status)
							break
						}
					}
					s, err := ParseSignature(signature)

					if err != nil {
						log.Printf("Error parsing signature: %s", err)
						break
					}

					d := newDecoder(bytes.NewReader(b), binary.LittleEndian)
					res, err := d.Decode(s)

					if err != nil {
						log.Printf("Error decoding message [%+v] : %s", b, err)
						break
					}

					log.Printf("Received application alljoyn message, signature: %s, bytes: %+v, decoded: %+v", signature, b, res)

					objPath := C.GoString(C.Get_AJ_Message_objPath())
					member := C.GoString(C.Get_AJ_Message_member())
					destination := C.GoString(C.Get_AJ_Message_destination())
					iface := C.GoString(C.Get_AJ_Message_iface())

					log.Printf("Message [objPath, member, iface, destination]: %s, %s, %s, %s", objPath, member, iface, destination)

					for _, service := range a.services[dbusService] {
						if (service.allJoynPath == objPath) && (service.allJoynService == destination) {
							log.Print("Found matching dbus service: %+v", service)
							remote := a.bus.Object(dbusService, dbus.ObjectPath(service.dbusPath))
							res := remote.Call(iface+"."+member, 0, res...)

							if res.Err != nil {
								log.Printf("Error calling dbus method (%s): %s", iface+"."+member, res.Err)
								return dbus.NewError(res.Err.Error(), nil)
							}

							C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
							buf := new(bytes.Buffer)
							enc := newEncoder(buf, binary.LittleEndian)
							err = enc.Encode(res.Body...)
							if err != nil {
								log.Printf("Error encoding result: %s", err)
								break
							}
							log.Printf("Encoded reply: %+v", buf.Bytes())
							C.AJ_DeliverMsgPartial((*C.AJ_Message)(reply), C.uint32_t(buf.Len()))
							C.AJ_MarshalRaw((*C.AJ_Message)(reply), unsafe.Pointer(&buf.Bytes()[0]), C.size_t(buf.Len()))
							C.AJ_DeliverMsg((*C.AJ_Message)(reply))
							break
						}
					}
				}
			default:
				/* Pass to the built-in handlers. */
				log.Printf("Passing msgId %+v to AllJoyn", msgId)
				status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
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
	bus.RequestName("com.devicehive.alljoyn.bridge",
		dbus.NameFlagDoNotQueue)

	if err != nil {
		log.Panic(err)
	}

	allJoynBridge := NewAllJoynBridge(bus, func(dbusService, dbusPath string) (*introspect.Node, error) {
		return introspect.Call(bus.Object(dbusService, dbus.ObjectPath(dbusPath)))
	})

	bus.Export(allJoynBridge, "/com/devicehive/alljoyn/bridge", "com.devicehive.alljoyn.bridge")
	select {}
}
