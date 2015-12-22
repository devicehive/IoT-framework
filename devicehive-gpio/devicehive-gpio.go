package main

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
	"strings"
)

const (
	BusName      = "com.devicehive.gpio"
	ErrorName    = "com.devicehive.Error"
	ServicePath  = "/com/devicehive/gpio"
	ServiceIFace = "com.devicehive.gpio.Service"
)

type Service struct {
	pins map[string]*Pin
}

// get all pins created
func (s *Service) GetAllPins() (map[string]string, *dbus.Error) {
	pins := make(map[string]string)
	for pin, p := range s.pins {
		pins[pin] = p.port
	}
	return pins, nil
}

// create new pin/port pair
func (s *Service) CreatePin(pin string, port string) *dbus.Error {
	// ensure pin/port not used
	for k, p := range s.pins {
		if k == pin {
			return newDBusError(fmt.Sprintf("Pin:%q already exists", pin))
		}
		if p.port == port {
			return newDBusError(fmt.Sprintf("Port:%q already used by pin:%q", port, k))
		}
	}

	s.pins[pin] = &Pin{port: port}
	//	self.InterfacesAdded(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})
	//	self.m_pin_services[pin].onPinChange += self.pin_changed_handler
	log.Printf("Create pin:%q for port:%q", pin, port)
	return nil // OK
}

// create multiple pins
func (s *Service) CreatePins(pins map[string]string) *dbus.Error {
	for pin, port := range pins {
		err := s.CreatePin(pin, port)
		if err != nil {
			return err
		}
	}
	return nil // OK
}

// delete pin
func (s *Service) DeletePin(pin string) *dbus.Error {
	p, ok := s.pins[pin]
	if !ok {
		return newDBusError(fmt.Sprintf("Pin:%q does not exist", pin))
	}
	log.Printf("Remove pin:%q for port:%q", pin, p.port)
	//	self.m_pin_services[pin].deinit()
	//	self.m_pin_services[pin].remove_from_connection()
	//	self.InterfacesRemoved(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})

	delete(s.pins, pin)
	return nil // OK
}

// delete all pins
func (s *Service) DeleteAllPins() *dbus.Error {
	for pin := range s.pins {
		err := s.DeletePin(pin)
		if err != nil {
			return err
		}
	}
	return nil // OK
}

type Pin struct {
	port string
}

// create new D-Bus connection
func newDBus(name string) (conn *dbus.Conn, err error) {
	// TODO: pass address and select one of system/session/custom

	conn, err = dbus.SystemBus()
	if err != nil {
		return
	}

	reply, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil {
		return
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		err = fmt.Errorf("D-Bus %q name already taken", name)
	}

	return
}

// create new D-Bus error
func newDBusError(body ...interface{}) *dbus.Error {
	return dbus.NewError(ErrorName, body)
}

// export Service object
func (s *Service) export(conn *dbus.Conn) (err error) {
	// export service itself
	err = conn.Export(s, ServicePath, ServiceIFace)
	if err != nil {
		return
	}

	// main service interface
	iface := introspect.Interface{
		Name:    ServiceIFace,
		Methods: introspect.Methods(s),
		Signals: []introspect.Signal{
			{
				Name: "CommandReceived",
				Args: []introspect.Arg{
					{"id", "t", "out"},
					{"name", "s", "out"},
					{"parameters", "s", "out"},
				},
			},
		},
	}

	// service
	node := introspect.Node{
		Interfaces: []introspect.Interface{iface},
	}
	err = conn.Export(introspect.NewIntrospectable(&node),
		ServicePath, introspect.IntrospectData.Name)
	if err != nil {
		return
	}

	// root object
	root := introspect.Node{
		Children: []introspect.Node{
			{Name: strings.TrimPrefix(ServicePath, "/")},
		},
	}
	conn.Export(introspect.NewIntrospectable(&root),
		"/", introspect.IntrospectData.Name)
	if err != nil {
		return
	}

	return // OK
}

// export Pin object
func exportPin(conn *dbus.Conn, pin *Pin) (err error) {
	return
}

// daemon entry point
func main() {
	bus, err := newDBus(BusName)
	if err != nil {
		log.Fatalf("no D-Bus access: %s", err)
	}

	s := &Service{}
	s.pins = make(map[string]*Pin)
	err = s.export(bus)
	if err != nil {
		log.Fatalf("failed to export service: %s", err)
	}

	select {}
}
