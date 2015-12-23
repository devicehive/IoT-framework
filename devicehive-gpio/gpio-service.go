package main // TODO: move to special package?

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
	"strings"
)

const (
	ErrorName    = "com.devicehive.Error"
	ServicePath  = "/com/devicehive/gpio"
	ServiceIFace = "com.devicehive.gpio.Service"
	ManagerIFace = "org.freedesktop.DBus.ObjectManager"
)

// main GPIO service
type Service struct {
	pins     map[string]*Pin
	dbusConn *dbus.Conn // not nil if exported to D-Bus
}

// create new Service
func NewService() *Service {
	s := &Service{}
	s.pins = make(map[string]*Pin)
	return s
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

	// introspection
	node := introspect.Node{
		Interfaces: []introspect.Interface{iface},
	}
	err = conn.Export(introspect.NewIntrospectable(&node),
		ServicePath, introspect.IntrospectData.Name)
	if err != nil {
		return
	}

	// export pins (first time only)
	if s.dbusConn == nil {
		for _, p := range s.pins {
			p.export(conn)
		}
	}

	// export root object
	err = s.exportRoot(conn)
	if err != nil {
		return
	}

	s.dbusConn = conn
	return // OK
}

// export root object
func (s *Service) exportRoot(conn *dbus.Conn) (err error) {
	// introspection
	root := introspect.Node{
		Children: []introspect.Node{
			{Name: strings.TrimPrefix(ServicePath, "/")},
		},
	}

	// all child pins
	for _, p := range s.pins {
		path := strings.TrimPrefix(string(p.dbusPath), "/")
		node := introspect.Node{Name: path}
		root.Children = append(root.Children, node)
	}

	// export root
	conn.Export(nil, "/", introspect.IntrospectData.Name) // unexport
	err = conn.Export(introspect.NewIntrospectable(&root),
		"/", introspect.IntrospectData.Name)
	if err != nil {
		return
	}

	return // OK
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
func (s *Service) createPin(pin string, port string) error {
	// ensure pin/port not used
	for k, p := range s.pins {
		if k == pin {
			return fmt.Errorf("Pin:%q already exists", pin)
		}
		if p.port == port {
			return fmt.Errorf("Port:%q already used by pin:%q", port, k)
		}
	}

	p := NewPin(pin, port)
	//	self.m_pin_services[pin].onPinChange += self.pin_changed_handler

	log.Printf("Create pin:%q for port:%q", pin, port)
	if s.dbusConn != nil {
		p.export(s.dbusConn)

		// emit ObjectManager.InterfacesAdded
//	self.InterfacesAdded(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})
//		s.dbusConn.Emit(ServicePath, fmt.Sprintf("%s.InterfacesAdded", ManagerIFace),
//			p.dbusPath, )
	}
	s.pins[pin] = p
	return nil // OK
}

// create new pin/port pair
func (s *Service) CreatePin(pin string, port string) *dbus.Error {
	err := s.createPin(pin, port)
	if err != nil {
		return newDBusError(err.Error())
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}

// create multiple pins
func (s *Service) CreatePins(pins map[string]string) *dbus.Error {
	for pin, port := range pins {
		err := s.createPin(pin, port)
		if err != nil {
			return newDBusError(err.Error())
		}
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}

// delete pin
func (s *Service) deletePin(pin string) error {
	p, ok := s.pins[pin]
	if !ok {
		return fmt.Errorf("Pin:%q does not exist", pin)
	}
	log.Printf("Remove pin:%q for port:%q", pin, p.port)
	p.Stop() // ignore error!
	//	self.m_pin_services[pin].remove_from_connection()

	if s.dbusConn != nil {
		p.unexport(s.dbusConn)
		
		// emit ObjectManager.InterfacesRemoved
	//	self.InterfacesRemoved(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})
//		s.dbusConn.Emit(ServicePath, fmt.Sprintf("%s.InterfacesRemoved", ManagerIFace),
//			p.dbusPath, )
	}

	delete(s.pins, pin)
	return nil // OK
}

// delete pin
func (s *Service) DeletePin(pin string) *dbus.Error {
	err := s.deletePin(pin)
	if err != nil {
		return newDBusError(err.Error())
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}

// delete all pins
func (s *Service) DeleteAllPins() *dbus.Error {
	for pin := range s.pins {
		err := s.deletePin(pin)
		if err != nil {
			return newDBusError(err.Error())
		}
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}
