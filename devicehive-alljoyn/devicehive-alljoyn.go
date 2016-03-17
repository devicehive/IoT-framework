package main

/*
#cgo CFLAGS: -DNDEBUG -I${SRCDIR}/lib -I${SRCDIR}/alljoyn/services/base_tcl/src -I${SRCDIR}/alljoyn/core/ajtcl/inc -I${SRCDIR}/alljoyn/core/ajtcl/dist/include -I${SRCDIR}/alljoyn/core/ajtcl/target/linux -I${SRCDIR}/alljoyn/services/base_tcl/notification/inc -I${SRCDIR}/alljoyn/services/base_tcl/services_common/inc -I${SRCDIR}/alljoyn/services/base_tcl/notification/src -I${SRCDIR}/alljoyn/services/base_tcl/services_common/src  -I${SRCDIR}/alljoyn/services/base_tcl/sample_apps/AppsCommon/inc -I${SRCDIR}/alljoyn/services/base_tcl/sample_apps/AppsCommon/src
#cgo LDFLAGS: -L${SRCDIR}/alljoyn/core/ajtcl -lajtcl
#include "cfuncs.h"
*/
import "C"

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/devicehive/IoT-framework/devicehive-alljoyn/ajmarshal"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const UNMARSHAL_TIMEOUT = 100
const PORT = 900
const AJ_CP_PORT = 1000
const AJ_ABOUT_IF = "org.alljoyn.About"
const AJ_ABOUT_GETABOUTDATA = "org.alljoyn.About.GetAboutData"

const AJ_NOTIFICATION_IF = "org.alljoyn.Notification"
const AJ_NOTIFICATION_MSG = "org.alljoyn.Notification.Notify"

const AJ_CP_PREFIX = "org.alljoyn.ControlPanel."
const AJ_CP_INTERFACE = "org.alljoyn.ControlPanel.ControlPanel"

const AJ_PROP_IF = "org.freedesktop.Dbus.Properties"

const ANN_SIGNAL_NAME = "com.devicehive.alljoyn.signal"
const ANN_SIGNAL_SESSIONLESS = "sessionless"

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

/* Gloabal variables to preserve from garbage collection to pass safely to cgo */
var interfaces []C.AJ_InterfaceDescription
var myPropGetterFunction PropGetterFunction
var myMessenger *AllJoynMessenger
var aboutData map[string]dbus.Variant

type PropGetterFunction func(reply *C.AJ_Message, language *C.char) C.AJ_Status

type MemberDescriptionProvider func(objIdx int, ifIdx int, memberIdx int, argIdx int, lang string) string

var memberDescriptionProvider MemberDescriptionProvider

type AllJoynBindingInfo struct {
	allJoynPath    string
	dbusPath       string
	introspectData *introspect.Node
}

type AllJoynServiceInfo struct {
	allJoynService    string
	dbusService       string // unique bus id name
	dbusServiceName   string // friendly name
	objects           []*AllJoynBindingInfo
	registeredObjects []*AllJoynBindingInfo
}

type AllJoynBridge struct {
	bus            *dbus.Conn
	signals        chan *dbus.Signal
	services       map[string]*AllJoynServiceInfo
	sessions       []uint32
	signalCache    map[string]*introspect.Signal
	signalIdxCache map[string]uint32
}

type AllJoynMessenger struct {
	dbusService string
	bus         *dbus.Conn
	binding     []*AllJoynBindingInfo
}

var (
	spawnUUID           = ""
	spawnDbusServiceId  = ""
	spawnDbusService    = ""
	spawnDbusPath       = ""
	spawnAlljoynService = ""
)

// initialize test environment
func init() {
	flag.StringVar(&spawnUUID, "spawn-uuid", "", "(do not use directly)")
	flag.StringVar(&spawnDbusServiceId, "spawn-dbus-service-id", "", "(do not use directly)")
	flag.StringVar(&spawnDbusService, "spawn-dbus-service", "", "(do not use directly)")
	flag.StringVar(&spawnDbusPath, "spawn-dbus-path", "", "(do not use directly)")
	flag.StringVar(&spawnAlljoynService, "spawn-alljoyn-service", "", "(do not use directly)")
	flag.Parse()
}

func NewAllJoynMessenger(dbusService string, bus *dbus.Conn, binding []*AllJoynBindingInfo) *AllJoynMessenger {
	return &AllJoynMessenger{dbusService, bus, binding}
}

func NewAllJoynBridge(bus *dbus.Conn) *AllJoynBridge {
	bridge := new(AllJoynBridge)
	bridge.bus = bus
	bridge.signals = make(chan *dbus.Signal, 100)
	bridge.services = make(map[string]*AllJoynServiceInfo)
	bridge.sessions = []uint32{}
	bridge.signalCache = make(map[string]*introspect.Signal)
	bridge.signalIdxCache = make(map[string]uint32)

	sbuffer := make(chan *dbus.Signal, 100)
	go bridge.signalsPump(sbuffer)
	bus.Signal(sbuffer)

	return bridge
}

func (a *AllJoynBridge) signalsPump(buffer chan *dbus.Signal) {
	for signal := range buffer {
		a.signals <- signal
	}
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
			// log.Print(methogString + argString)
			desc = append(desc, C.CString(methogString+argString))
		}

		for _, signal := range iface.Signals {

			var prefix string

			if isSessionless(&signal) {
				prefix = "!&"
			} else {
				prefix = "!"
			}

			signalString := prefix + signal.Name
			argString := ParseArguments(signal.Args)
			// log.Print(signalString + argString)
			desc = append(desc, C.CString(signalString+argString))
		}

		for _, prop := range iface.Properties {
			propString := "@" + ParseArgumentOrProperty(prop.Name, prop.Access, prop.Type)
			// log.Print(propString)
			desc = append(desc, C.CString(propString))
		}

		desc = append(desc, nil)
		//		log.Println(desc)
		res = append(res, (C.AJ_InterfaceDescription)(&desc[0]))
	}
	res = append(res, nil)
	return res
}

func getAllJoynObjectFlags(service *AllJoynBindingInfo) C.uint8_t {

	var hasControlPanel = false
	var hasOtherCPInterfaces = false

	for _, iface := range service.introspectData.Interfaces {
		hasOtherCPInterfaces = hasOtherCPInterfaces || strings.HasPrefix(iface.Name, AJ_CP_PREFIX)
		hasControlPanel = hasControlPanel || (iface.Name == AJ_CP_INTERFACE)
	}

	if hasOtherCPInterfaces {
		if hasControlPanel {
			return C.AJ_OBJ_FLAG_ANNOUNCED
		} else {
			return C.uint8_t(0) //C.AJ_OBJ_FLAG_HIDDEN
		}
	} else {
		return C.AJ_OBJ_FLAG_ANNOUNCED
		// it has to be C.AJ_OBJ_FLAG_DESCRIBED
		// but not yet compatible with AJ Explorer #151
	}
}

func hasInterface(object *AllJoynBindingInfo, ifname string) bool {
	for _, iface := range object.introspectData.Interfaces {
		if iface.Name == ifname {
			return true
		}
	}
	return false
}

//export PutMemberDescription
func PutMemberDescription(objIdx C.uint32_t, ifIdx C.uint32_t,
	memberIdx C.uint32_t, argIdx C.uint32_t,
	language *C.char, dest *C.char, maxlen C.uint32_t) {

	desc := memberDescriptionProvider(int(objIdx), int(ifIdx), int(memberIdx), int(argIdx), C.GoString(language))

	if len(desc) >= int(maxlen) {
		desc = desc[0:int(maxlen)]
	}

	buff := C.CString(desc)
	defer C.free(unsafe.Pointer(buff))
	C.strcpy(dest, buff)
	//	for i := 0; i < len(desc) && i < int(maxlen); i++ {
	//		dest  = C.char(desc[i])
	//	}
}

func GetKnownObjects(objects []*AllJoynBindingInfo) (*AllJoynBindingInfo, *AllJoynBindingInfo, []*AllJoynBindingInfo) {

	other := make([]*AllJoynBindingInfo, 0, len(objects))

	var notificationsObj, aboutObj *AllJoynBindingInfo

	for _, object := range objects {
		if hasInterface(object, AJ_NOTIFICATION_IF) {
			notificationsObj = object
		} else if hasInterface(object, AJ_ABOUT_IF) {
			aboutObj = object
		} else {
			other = append(other, object)
		}
	}
	return notificationsObj, aboutObj, other
}

func GetAllJoynObjects(objects []*AllJoynBindingInfo) *C.AJ_Object {
	array := C.Allocate_AJ_Object_Array(C.uint32_t(len(objects) + 1))

	for i, object := range objects {
		interfaces = ParseAllJoynInterfaces(object.introspectData.Interfaces)
		flags := getAllJoynObjectFlags(object)
		log.Printf("Creating Object %s %b", object.introspectData.Name, flags)
		C.Create_AJ_Object(C.uint32_t(i), array, C.CString(object.introspectData.Name), &interfaces[0], flags, unsafe.Pointer(nil))
	}
	C.Create_AJ_Object(C.uint32_t(len(objects)+1), array, nil, nil, 0, nil)
	return array
}

////export MyAboutPropGetter
//func MyAboutPropGetter(reply *C.AJ_Message, language *C.char) C.AJ_Status {
//	return myMessenger.MyAboutPropGetter_member(reply, language)
//}

//func (m *AllJoynMessenger) MyAboutPropGetter_member(reply *C.AJ_Message, language *C.char) C.AJ_Status {
//	// log.Printf("MyAboutPropGetter_member(): %+v", m)
//	aboutInterfacePath := m.binding[0].dbusPath

//	log.Printf("About interface path: %s", aboutInterfacePath)
//	err := m.callRemoteMethod(reply, aboutInterfacePath, "org.alljoyn.About.GetAboutData", []interface{}{C.GoString(language)})
//	if err != nil {
//		log.Printf("Error calling org.alljoyn.About for [%+v]: %s", m.binding, err)
//		return C.AJ_ERR_NO_MATCH
//	}

//	// log.Printf("About message signature, offset: %s, %d", C.GoString(reply.signature), reply.sigOffset)
//	return C.AJ_OK
//}

func (m *AllJoynMessenger) callRemoteMethod(message *C.AJ_Message, path, member string, arguments []interface{}) (err error) {
	log.Printf("MSG: message.hdr=%p", message.hdr)
	remote := m.bus.Object(m.dbusService, dbus.ObjectPath(path))
	// log.Printf("%s Argument[0] %+v", member, reflect.ValueOf(arguments[0]).Type())
	res := remote.Call(member, 0, arguments...)

	if res.Err != nil {
		log.Printf("Error calling dbus method (%s): %s", member, res.Err)
		return res.Err
	}

	//pad := 4 - (int)(message.bodyBytes)%4
	//C.WriteBytes((*C.AJ_Message)(message), nil, (C.size_t)(0), (C.size_t)(pad))

	buf := new(bytes.Buffer)
	buf.Write(make([]byte, (int)(message.bodyBytes)))
	enc := ajmarshal.NewEncoderAtOffset(buf, (int)(message.bodyBytes), binary.LittleEndian)
	pad, err := enc.Encode(res.Body...)
	// log.Printf("Padding of the encoded buffer: %d", pad)
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
	// log.Printf("Length before: %d", message.bodyBytes)

	newBuf := buf.Bytes()[(int)(message.bodyBytes)+pad:]
	//log.Printf("Buffer to write into AllJoyn: %+v, %d", newBuf, len(newBuf))
	//	hdr := message.hdr
	//	if hdr.flags&C.uint8_t(C.AJ_FLAG_ENCRYPTED) == 0 {
	//		hdr = nil
	//	} else {
	//		message.hdr.flags &= ^C.uint8_t(C.AJ_FLAG_ENCRYPTED)
	//	}

	if len(newBuf) > 0 {
		log.Printf("MSG: message.hdr=%p", message.hdr)
		C.AJ_DeliverMsgPartial((*C.AJ_Message)(message), C.uint32_t(len(newBuf)))
		log.Printf("MSG: message.hdr=%p", message.hdr)

		C.AJ_MarshalRaw((*C.AJ_Message)(message), unsafe.Pointer(&newBuf[0]), C.size_t(len(newBuf)))
	} else {
		C.AJ_MarshalRaw((*C.AJ_Message)(message), unsafe.Pointer(&newBuf), C.size_t(0))
	}
	//log.Printf("New buff reply, len: %+v, %d", newBuf, len(newBuf))
	//	if hdr != nil {
	//		message.hdr = hdr
	//		message.hdr.flags &= ^C.uint8_t(C.AJ_FLAG_ENCRYPTED)
	//	}
	log.Printf("MSG: message.hdr=%p", message.hdr)
	return nil
}

func safeString(p *C.char) string {

	if p == nil || unsafe.Pointer(p) == unsafe.Pointer(nil) {
		return ""
	} else {
		return C.GoString(p)
	}
}

func dumpMessage() string {

	objPath := safeString(C.Get_AJ_Message_objPath())
	member := safeString(C.Get_AJ_Message_member())
	destination := safeString(C.Get_AJ_Message_destination())
	iface := safeString(C.Get_AJ_Message_iface())

	return fmt.Sprintf("\tPath: %s \r\n\tMember: %s.%s \r\n\tDestination: %s", objPath, iface, member, destination)

}

func (m *AllJoynMessenger) forwardAllJoynMessage(msgId uint32) (err error) {
	log.Printf("Passing msgId 0x%x to DBus", msgId)
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
			// log.Printf("Reading RAW message, status = %d, actual = %d", status, actual)
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

	d := ajmarshal.NewDecoder(bytes.NewReader(b), binary.LittleEndian)
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

	log.Printf("Message [objPath, member, iface, destination]: %s, %s, %s, %s", objPath, member, iface, destination)

	for _, service := range m.binding {
		if service.allJoynPath == objPath {
			log.Printf("Found matching dbus service: %+v", service)
			C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
			m.callRemoteMethod((*C.AJ_Message)(reply), service.dbusPath, iface+"."+member, res)
			C.AJ_DeliverMsg((*C.AJ_Message)(reply))
			break
		}
	}
	return nil
}

func SubscribeToSignals(bus *dbus.Conn, service *AllJoynServiceInfo) {
	for _, obj := range service.objects {
		query := "type='signal',sender='" + service.dbusService + "',path='" + obj.dbusPath + "'"
		bus.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, query)
		// log.Printf("FILTERING SIGNALS: %s", query)
	}
}

func (a *AllJoynBridge) addSession(sessionId uint32) {
	a.removeSession(sessionId)
	a.sessions = append(a.sessions, sessionId)
}

func (a *AllJoynBridge) removeSession(sessionId uint32) {
	for i, sid := range a.sessions {
		if sid == sessionId {
			a.sessions = append(a.sessions[:i], a.sessions[i+1:]...)
			return
		}
	}
}

func isSessionless(sig *introspect.Signal) bool {
	if sig == nil || len(sig.Annotations) == 0 {
		return false
	}

	for _, a := range sig.Annotations {
		if a.Name == ANN_SIGNAL_NAME && a.Value == ANN_SIGNAL_SESSIONLESS {
			return true
		}
	}

	return false
}

func (a *AllJoynBridge) findSignal(signal *dbus.Signal) (*introspect.Signal, uint32) {

	key := signal.Name + string(signal.Path) + signal.Sender

	// check for cached value
	if sig, ok := a.signalCache[key]; ok {
		return sig, a.signalIdxCache[key]
	}

	sepPos := strings.LastIndex(signal.Name, ".")
	singalInterface := signal.Name[:sepPos]
	singalName := signal.Name[sepPos+1:]

	log.Printf("**** Inteface: %s, Name: %s", singalInterface, singalName)

	for _, service := range a.services {
		if service.dbusService == signal.Sender {
			log.Printf("Found matching service: %s", service.dbusServiceName)

			// using registeredObjects here as index is calculated for registered alljoyn objects
			for objIdx, object := range service.registeredObjects {
				if object.dbusPath == string(signal.Path) {
					for ifIdx, iface := range object.introspectData.Interfaces {
						if iface.Name == singalInterface {
							for sIdx, sgn := range iface.Signals {
								if sgn.Name == singalName {

									memberIdx := len(iface.Methods) + sIdx
									signalMessageId := 0x01000000 | (uint32(objIdx) << 16) | (uint32(ifIdx) << 8) | uint32(memberIdx)
									log.Printf("##### Signal: %s => 0x%X  (%d %d %d)", signal.Name, signalMessageId, objIdx, ifIdx, memberIdx)

									// cache value for future use
									a.signalCache[key] = &sgn
									a.signalIdxCache[key] = signalMessageId

									return &sgn, signalMessageId
								}
							}
							// log.Printf("Could not find any matching signal for: %+v", signal)
							return nil, 0
						}
					}
					// log.Printf("Could not find any matching interface for: %+v", signal)
					return nil, 0
				}
			}
			// log.Printf("Could not find any matching obejct for: %+v", signal)
			return nil, 0
		}
	}
	// log.Printf("Could not find any matching service for: %+v", signal)
	return nil, 0
}

func parseNotificationBody(body []interface{}) (msgType uint16, lang string, msg string, success bool) {

	success = true

	if data, ok := body[2].(uint16); ok {
		msgType = data
	} else {
		log.Printf("Could not parse msgType: %+v", body[2])
		success = success && false
	}

	if message, ok := body[7].(map[string]string); ok {
		for key, value := range message {
			lang = key
			msg = value
		}
	} else {
		log.Printf("Could not parse message: %+v", body[7])
		success = success && false
	}

	return
}

func (a *AllJoynBridge) processNotificationSignal(signal *dbus.Signal) bool {

	if signal.Name != AJ_NOTIFICATION_MSG {
		return false
	}

	if msgType, lang, message, ok := parseNotificationBody(signal.Body); ok {
		log.Printf("NOTIFY: %v => [%s:%s]", msgType, lang, message)
		C.SendNotification(C.uint16_t(msgType), C.CString(lang), C.CString(message))
	} else {
		log.Printf("ERROR PARSING NOTIFY BODY: %+v", signal.Body)
	}

	return true
}

func (a *AllJoynBridge) processSignals() {

	// work on currently reveived signals only
	// TODO:https://golang.org/doc/effective_go.html#leaky_buffer
	signals := a.signals
	a.signals = make(chan *dbus.Signal, 100)
	close(signals)

	// log.Printf("Processing signals: %d", len(signals))

	for signal := range signals {
		log.Printf("**** Incoming Signal: %+v", signal)

		if a.processNotificationSignal(signal) {
			continue
		}

		// get signal signature
		sigIntrospect, msgId := a.findSignal(signal)

		if msgId == 0 {
			log.Printf("Could not find any matching service for signal: %+v", signal)
			continue
		}
		log.Printf("%v", sigIntrospect)
		if isSessionless(sigIntrospect) {
			log.Printf("SESSIONLESS SIGNAL!")

			var status C.AJ_Status = C.AJ_OK
			msg := C.Get_AJ_Message()

			status = C.AJ_MarshalSignal_cgo(msg, C.uint32_t(msgId), C.uint32_t(0), C.AJ_FLAG_SESSIONLESS, C.uint32_t(0))
			log.Printf("**** AJ_MarshalSignal: %s", status)

			if len(signal.Body) > 0 {

				buf := new(bytes.Buffer)
				buf.Write(make([]byte, (int)(msg.bodyBytes)))
				enc := ajmarshal.NewEncoderAtOffset(buf, (int)(msg.bodyBytes), binary.LittleEndian)
				pad, err := enc.Encode(signal.Body...)

				if err != nil {
					log.Printf("Error encoding result: %s", err)
					continue
				}

				newBuf := buf.Bytes()[(int)(msg.bodyBytes)+pad:]

				if len(newBuf) > 0 {
					status = C.AJ_DeliverMsgPartial((*C.AJ_Message)(msg), C.uint32_t(len(newBuf)))
					log.Printf("**** AJ_DeliverMsgPartial: %s", status)
					status = C.AJ_MarshalRaw((*C.AJ_Message)(msg), unsafe.Pointer(&newBuf[0]), C.size_t(len(newBuf)))
				} else {
					status = C.AJ_MarshalRaw((*C.AJ_Message)(msg), unsafe.Pointer(&newBuf), C.size_t(0))
				}
				log.Printf("**** AJ_MarshalRaw: %s", status)

			}

			status = C.AJ_DeliverMsg((*C.AJ_Message)(msg))
			log.Printf("**** AJ_DeliverMsg: %s", status)

			status = C.AJ_CloseMsg((*C.AJ_Message)(msg))
			log.Printf("**** AJ_CloseMsg: %s", status)

		} else {

			for _, sessionId := range a.sessions {

				var status C.AJ_Status = C.AJ_OK
				msg := C.Get_AJ_Message()

				status = C.AJ_MarshalSignal_cgo(msg, C.uint32_t(msgId), C.uint32_t(sessionId), C.uint8_t(0), C.uint32_t(0))
				log.Printf("**** AJ_MarshalSignal: %s", status)

				if len(signal.Body) > 0 {

					buf := new(bytes.Buffer)
					buf.Write(make([]byte, (int)(msg.bodyBytes)))
					enc := ajmarshal.NewEncoderAtOffset(buf, (int)(msg.bodyBytes), binary.LittleEndian)
					pad, err := enc.Encode(signal.Body...)

					if err != nil {
						log.Printf("Error encoding result: %s", err)
						continue
					}

					newBuf := buf.Bytes()[(int)(msg.bodyBytes)+pad:]

					if len(newBuf) > 0 {
						status = C.AJ_DeliverMsgPartial((*C.AJ_Message)(msg), C.uint32_t(len(newBuf)))
						log.Printf("**** AJ_DeliverMsgPartial: %s", status)
						status = C.AJ_MarshalRaw((*C.AJ_Message)(msg), unsafe.Pointer(&newBuf[0]), C.size_t(len(newBuf)))
					} else {
						status = C.AJ_MarshalRaw((*C.AJ_Message)(msg), unsafe.Pointer(&newBuf), C.size_t(0))
					}
					log.Printf("**** AJ_MarshalRaw: %s", status)

				}

				status = C.AJ_DeliverMsg((*C.AJ_Message)(msg))
				log.Printf("**** AJ_DeliverMsg: %s", status)

				status = C.AJ_CloseMsg((*C.AJ_Message)(msg))
				log.Printf("**** AJ_CloseMsg: %s", status)
			}
		}
	}
}

func (a *AllJoynBridge) fetchAboutData(svcInfo *AllJoynServiceInfo, objInfo *AllJoynBindingInfo) error {
	obj := a.bus.Object(svcInfo.dbusService, dbus.ObjectPath(objInfo.dbusPath))
	call := obj.Call(AJ_ABOUT_GETABOUTDATA, 0, "en")
	if call.Err != nil {
		log.Printf("Error calling %s: %v", AJ_ABOUT_GETABOUTDATA, call.Err)
		return call.Err
	}

	//	fill aboutData with values

	call.Store(&aboutData)

	log.Printf("ABOUT: %+v", aboutData)

	for key, value := range aboutData {
		//		log.Printf("%s(%s)", key, value.Signature())
		switch value.Signature().String() {
		case "ay":
			C.SetProperty(C.CString(key), unsafe.Pointer(C.CString(fmt.Sprintf("%x", value.Value()))))
		default:
			C.SetProperty(C.CString(key), unsafe.Pointer(C.CString(value.String())))

		}

	}

	return nil
}

func (status C.AJ_Status) String() string {
	return C.GoString(C.AJ_StatusText(status))
}

//func parseMemberId(id uint32) (objIdx int, ifIdx int, memberIdx int, argIdx int) {
//	objIdx = int((id >> 24) & 0xFF)
//	ifIdx = int((id >> 16) & 0xFF)
//	memberIdx = int((id >> 8) & 0xFF)
//	argIdx = int(id & 0xFF)
//	log.Printf("PARSE IDX(0x%X): %d, %d, %d, %d", id, objIdx, ifIdx, memberIdx, argIdx)
//	return
//}

func (service *AllJoynServiceInfo) MemberDescriptionProvider(objIdx int, ifIdx int, memberIdx int, argIdx int, lang string) string {

	obj := service.registeredObjects[objIdx].introspectData

	// zero position means obejct/interface/member own description rather than child position
	if ifIdx == 0 {
		log.Printf("MEMBER DESC (%v, %v, %v ,%v) => %s", objIdx, ifIdx, memberIdx, argIdx, obj.Name)
		return obj.Name
	}

	iface := obj.Interfaces[ifIdx-1]
	methods, signals, properties := len(iface.Methods), len(iface.Signals), len(iface.Properties)

	if memberIdx == 0 {
		log.Printf("MEMBER DESC (%v, %v, %v ,%v) => %s", objIdx, ifIdx, memberIdx, argIdx, iface.Name)
		return iface.Name
	} else {
		memberIdx = memberIdx - 1
	}

	var name string

	if memberIdx < methods {
		if argIdx == 0 {
			name = iface.Methods[memberIdx].Name
		} else {
			name = iface.Methods[memberIdx].Args[argIdx-1].Name
		}
	} else if memberIdx < methods+signals {
		if argIdx == 0 {
			name = iface.Signals[memberIdx-methods].Name
		} else {
			name = iface.Signals[memberIdx-methods].Args[argIdx-1].Name
		}
	} else if memberIdx < methods+signals+properties {
		name = iface.Properties[memberIdx-methods-signals].Name
	} else {
		name = fmt.Sprintf("Unknown %v, %v, %v ,%v", objIdx, ifIdx, memberIdx, argIdx)
	}

	log.Printf("MEMBER DESC (%v, %v, %v ,%v) => %s", objIdx, ifIdx, memberIdx, argIdx, name)

	return name
}

func (a *AllJoynBridge) startAllJoyn(uuid string) *dbus.Error {
	service := a.services[uuid]

	notificationsObj, aboutObj, otherObjects := GetKnownObjects(service.objects)

	objects := GetAllJoynObjects(otherObjects)
	service.registeredObjects = otherObjects

	memberDescriptionProvider = service.MemberDescriptionProvider

	a.fetchAboutData(service, aboutObj)

	SubscribeToSignals(a.bus, service)

	myMessenger = NewAllJoynMessenger(service.dbusService, a.bus, otherObjects)

	var status C.AJ_Status = C.AJ_OK

	connected := false
	busAttachment := C.Get_AJ_BusAttachment()

	C.AJ_Initialize()
	C.AJ_RegisterDescriptionLanguages((**C.char)(C.getLanguages()))
	C.AJ_AboutRegisterPropStoreGetter((C.AJ_AboutPropGetter)(unsafe.Pointer(C.MyAboutPropGetter)))
	C.AJ_RegisterObjectListWithDescriptions(objects, 1, (C.AJ_DescriptionLookupFunc)(C.MyTranslator))
	//	C.AJ_RegisterObjects(objects, nil)
	C.AJ_SetMinProtoVersion(10)

	// DEBUG
	// C.AJ_PrintXML(objects)
	// C.AJ_PrintXMLWithDescriptions(objects, C.CString("en"))

	if notificationsObj != nil {
		log.Println("NOTIFICATIONS ENABLED")
		//		C.InitNotificationContent()
		status = C.AJNS_Producer_Start()
	}

	if status != C.AJ_OK {
		log.Printf("Error: AJNS_Producer_Start()=> %s", status)
	}

	msg := C.Get_AJ_Message()
	C.AJ_ClearAuthContext()

	log.Printf("CreateAJ_BusAttachment(): %+v", busAttachment)

	go func() {
		for {
			if !connected {
				busNodeName := C.CString("org.alljoyn.BusNode")
				defer C.free(unsafe.Pointer(busNodeName))

				status = C.AJ_StartService(busAttachment,
					//					(*C.char)(busNodeName),
					(*C.char)(unsafe.Pointer(nil)),
					60*1000, // TODO: Move connection timeout to config
					C.FALSE,
					C.uint16_t(PORT), // TODO: Move port to config
					C.CString(service.allJoynService),
					C.AJ_NAME_REQ_DO_NOT_QUEUE,
					(*C.AJ_SessionOpts)(C.Get_Session_Opts()),
				)

				log.Printf("StartService returned %s", status)

				if status != C.AJ_OK {

					continue
				}

				connected = true

				// Start Control Panel by binding a session port
				status = C.AJ_BusBindSessionPort(busAttachment,
					AJ_CP_PORT, (*C.AJ_SessionOpts)(C.Get_Session_Opts()), 0)

				if status != C.AJ_OK {
					log.Printf(("Failed to send bind control panel port message"))
				}

			}

			status = C.AJ_UnmarshalMsg(busAttachment, msg, UNMARSHAL_TIMEOUT)

			if C.AJ_ERR_TIMEOUT == status {
				// no incoming messages, we can do our work
				a.processSignals()
				continue
			}

			log.Printf("----------------------------")
			log.Printf("AJ_UnmarshalMsg: %s", status)

			if C.AJ_OK == status {

				msgId := C.Get_AJ_Message_msgId()
				log.Printf("MSG ID: 0x%X\n%s", msgId, dumpMessage())

				switch {
				case msgId == C.AJ_METHOD_ACCEPT_SESSION:
					{

						var c_port C.uint16_t
						var c_sessionId C.uint32_t

						C.UnmarshalJoinSessionArgs((*C.AJ_Message)(msg), &c_port, &c_sessionId)

						port := int(c_port)
						sessionId := uint32(c_sessionId)

						if port == PORT || port == AJ_CP_PORT {
							status = C.AJ_BusReplyAcceptSession((*C.AJ_Message)(msg), C.TRUE)
							log.Printf("ACCEPT_SESSION: %d at %d port", sessionId, port)
							a.addSession(sessionId)
						} else {
							status = C.AJ_BusReplyAcceptSession((*C.AJ_Message)(msg), C.FALSE)
							log.Printf("REJECT_SESSION: %d at %d port", sessionId, port)
						}
					}

				case msgId == C.AJ_SIGNAL_SESSION_LOST_WITH_REASON:
					{
						var c_sessionId, c_reason C.uint32_t

						C.UnmarshalLostSessionArgs((*C.AJ_Message)(msg), &c_sessionId, &c_reason)

						sessionId := uint32(c_sessionId)
						reason := uint32(c_reason)
						log.Printf("Session lost: %d as of reason %d", sessionId, reason)
						a.removeSession(sessionId)
					}
					//				case uint32(msgId) == 0x1010003: // Config.GetConfigurations
					//					// our forwardAllJoyn doesn't support encrypted messages which config service is,
					//					// so we handle it here manually
					//					{
					//						//C.AJ_UnmarshalArgs(msg, "s", &language);
					//						reply := C.Get_AJ_ReplyMessage()
					//						C.AJ_MarshalReplyMsg((*C.AJ_Message)(msg), (*C.AJ_Message)(reply))
					//						C.AJ_MarshalContainer((*C.AJ_Message)(reply), (*C.AJ_Arg)(C.Get_Arg()), C.AJ_ARG_ARRAY)
					//						C.AJ_MarshalArgs_cgo((*C.AJ_Message)(reply), C.CString("{sv}"), C.CString("DeviceName"), C.CString("s"), C.CString("DeviceHiveVB"))
					//						C.AJ_MarshalArgs_cgo((*C.AJ_Message)(reply), C.CString("{sv}"), C.CString("DefaultLanguage"), C.CString("s"), C.CString("en"))
					//						C.AJ_MarshalCloseContainer((*C.AJ_Message)(reply), (*C.AJ_Arg)(C.Get_Arg()))
					//						C.AJ_DeliverMsg((*C.AJ_Message)(reply))
					//					}
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
						log.Printf("Passing msgId 0x%x to AllJoyn", uint32(msgId))
						status = C.AJ_BusHandleBusMessage((*C.AJ_Message)(msg))
						log.Printf("AllJoyn returned %s", status)
					}
				}

				// Any received packets indicates the link is active, so call to reinforce the bus link state
				C.AJ_NotifyLinkActive()
			}

			/* Messages MUST be discarded to free resources. */
			C.AJ_CloseMsg((*C.AJ_Message)(msg))

			// if status == C.AJ_OK {
			// 	log.Print("***C.AJ_AboutAnnounce***")
			// 	C.AJ_AboutAnnounce(busAttachment)
			// }

			if status == C.AJ_ERR_READ {
				C.AJ_Disconnect(busAttachment)
				log.Print("AllJoyn disconnected, retrying")
				connected = false
				C.AJ_Sleep(1000 * 2) // TODO: Move sleep time to const
			}
		}
	}()
	return nil
}

func traverseDbusObjects(bus *dbus.Conn, dbusService, dbusPath string, fn func(path string, node *introspect.Node)) {
	var xmldata string
	var node introspect.Node

	var o = bus.Object(dbusService, dbus.ObjectPath(dbusPath))
	err := o.Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&xmldata)

	if err != nil {
		log.Printf("Error getting introspect from [%s, %s]: %s", dbusService, dbusPath, err)
	}

	err = xml.NewDecoder(strings.NewReader(xmldata)).Decode(&node)
	if err != nil {
		log.Printf("Error decoding introspect from [%s, %s]: %s", dbusService, dbusPath, err)
	}
	// log.Printf("Introspect: %+v", node)

	if node.Name != "" && len(node.Interfaces) > 0 {
		fn(dbusPath, &node)
	}

	for _, child := range node.Children {
		traverseDbusObjects(bus, dbusService, dbusPath+"/"+child.Name, fn)
	}
}

func (a *AllJoynBridge) AddService(dbusService, dbusPath, alljoynService string, sender dbus.Sender) (string, *dbus.Error) {
	// generate unique UUID
	var uuid string
	var err error
	for {
		uuid, err = newUUID()
		if err != nil {
			log.Printf("Error: %v", err)
			return "", dbus.NewError("com.devicehive.Error", []interface{}{err.Error})
		}

		if _, exists := a.services[uuid]; !exists {
			break // UUID is unique, done
		}

		// otherwise generate new one at next iteration...
	}

	// a.addService(uuid, string(sender), dbusService, dbusPath, alljoynService)
	async := exec.Command(os.Args[0],
		"--spawn-uuid", uuid,
		"--spawn-dbus-service-id", string(sender),
		"--spawn-dbus-service", dbusService,
		"--spawn-dbus-path", dbusPath,
		"--spawn-alljoyn-service", alljoynService)
	async.Stdout = os.Stdout
	async.Stderr = os.Stderr
	async.Start()
	// TODO: add 'async' to the list for management purposes

	return uuid, nil
}

func (a *AllJoynBridge) addService(uuid, dbusServiceId, dbusService, dbusPath, alljoynService string) {
	log.Printf("Traversing objects tree for %s (%s [%s] at %s):", uuid, dbusService, dbusServiceId, dbusPath)

	var bindings []*AllJoynBindingInfo
	traverseDbusObjects(a.bus, dbusServiceId, dbusPath, func(path string, node *introspect.Node) {
		allJoynPath := strings.TrimPrefix(path, dbusPath)
		bindings = append(bindings, &AllJoynBindingInfo{allJoynPath, path, node})
		log.Printf("Found Object: %s with %d interfaces", allJoynPath, len(node.Interfaces))
	})

	a.services[uuid] = &AllJoynServiceInfo{alljoynService, dbusServiceId, dbusService, bindings, nil}

	log.Printf("Added %s service with %d AJ objects", alljoynService, len(bindings))

	if len(bindings) != 0 {
		go a.startAllJoyn(uuid)
	}
}

func main() {
	bus, err := dbus.SystemBus()

	if err != nil {
		log.Fatal(err)
	}

	// run as a child
	if len(spawnUUID) != 0 && len(spawnDbusServiceId) != 0 && len(spawnDbusPath) != 0 {
		allJoynBridge := NewAllJoynBridge(bus)
		allJoynBridge.addService(spawnUUID, spawnDbusServiceId, spawnDbusService, spawnDbusPath, spawnAlljoynService)
		select {} // exit?
		return
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
				Name: "com/devicehive/alljoyn/bridge",
			},
		},
	}

	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/alljoyn/bridge", "org.freedesktop.DBus.Introspectable")
	bus.Export(introspect.NewIntrospectable(root), "/", "org.freedesktop.DBus.Introspectable") // workaroud for dbus issue #14

	log.Printf("Bridge is Running.")

	select {}
}
