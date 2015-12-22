package main

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"io/ioutil"
	"flag"
	"log"
	"strings"
	"os"
)

const (
	BusName      = "com.devicehive.gpio"
	ErrorName    = "com.devicehive.Error"
	ServicePath  = "/com/devicehive/gpio"
	ServiceIFace = "com.devicehive.gpio.Service"
	PinPathBase  = ServicePath
	PinIFace     = "com.devicehive.gpio.Pin"
)

// main service
type Service struct {
	pins map[string]*Pin
	dbusConn *dbus.Conn // not nil if exported
}

// create new Service object
func newService() *Service {
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

	// pins (first time only)
	if s.dbusConn == nil {
		for _, p := range s.pins {
			p.export(conn)
		}
	}

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
	for _, p := range s.pins {
		path := strings.TrimPrefix(string(p.dbusPath), "/")
		node := introspect.Node{Name: path}
		root.Children = append(root.Children, node)
	}
	conn.Export(nil, "/", introspect.IntrospectData.Name) // unexport
	conn.Export(introspect.NewIntrospectable(&root),
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
func (s *Service) createPin(pin string, port string) *dbus.Error {
	// ensure pin/port not used
	for k, p := range s.pins {
		if k == pin {
			return newDBusError(fmt.Sprintf("Pin:%q already exists", pin))
		}
		if p.port == port {
			return newDBusError(fmt.Sprintf("Port:%q already used by pin:%q", port, k))
		}
	}

	p := newPin(pin, port)
	//	self.InterfacesAdded(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})
	//	self.m_pin_services[pin].onPinChange += self.pin_changed_handler

	log.Printf("Create pin:%q for port:%q", pin, port)
	if s.dbusConn != nil {
		p.export(s.dbusConn)
	}
	s.pins[pin] = p
	return nil // OK
}

// create new pin/port pair
func (s *Service) CreatePin(pin string, port string) *dbus.Error {
	err := s.createPin(pin, port)
	if err != nil {
		return err
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
			return err
		}
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}

// delete pin
func (s *Service) deletePin(pin string) *dbus.Error {
	p, ok := s.pins[pin]
	if !ok {
		return newDBusError(fmt.Sprintf("Pin:%q does not exist", pin))
	}
	log.Printf("Remove pin:%q for port:%q", pin, p.port)
	//	self.m_pin_services[pin].deinit()
	//	self.m_pin_services[pin].remove_from_connection()
	//	self.InterfacesRemoved(self.m_pin_services[pin].m_service_path, {DBUS_PIN_INTERFACE: dict()})

	if s.dbusConn != nil {
		p.unexport(s.dbusConn)
	}

	delete(s.pins, pin)
	return nil // OK
}

// delete pin
func (s *Service) DeletePin(pin string) *dbus.Error {
	err := s.deletePin(pin)
	if err != nil {
		return err
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
			return err
		}
	}

	// update root object
	if s.dbusConn != nil {
		s.exportRoot(s.dbusConn)
	}
	return nil // OK
}

// exported pin
type Pin struct {
	port string
	dbusPath dbus.ObjectPath

	portDir string // port directory path
	portVal string // value file

	attached bool
}

// create new pin
func newPin(pin, port string) *Pin {
	p := &Pin{}
	p.port = port
	//p.pin = pin
	p.dbusPath = dbus.ObjectPath(fmt.Sprintf("%s/%s", PinPathBase, pin))
	if p.isAnalog() {
		p.portDir = fmt.Sprintf("/sys/bus/iio/devices/iio:device0/in_voltage%s_raw", port[3:]) // skip AIN prefix
		p.portVal = p.portDir
	} else {
		p.portDir = fmt.Sprintf("/sys/class/gpio/gpio%s", port)
		p.portVal = fmt.Sprintf("%s/value", p.portDir)
	}
	return p
}


// export Pin object
func (p *Pin) export(conn *dbus.Conn) (err error) {
	// export pin itself
	err = conn.Export(p, p.dbusPath, PinIFace)
	if err != nil {
		return
	}

	// main pin interface
	iface := introspect.Interface{
		Name:    PinIFace,
		Methods: introspect.Methods(p),
		Signals: []introspect.Signal{
			{
				Name: "ValueChanged",
				Args: []introspect.Arg{
					{"pin", "s", "out"},
					{"value", "s", "out"},
				},
			},
		},
	}

	// introspection
	node := introspect.Node{
		Interfaces: []introspect.Interface{iface},
	}
	err = conn.Export(introspect.NewIntrospectable(&node),
		p.dbusPath, introspect.IntrospectData.Name)
	if err != nil {
		return
	}

	return // OK
}

// unexport Pin object
func (p *Pin) unexport(conn *dbus.Conn) (err error) {
	// unexport pin itself
	err = conn.Export(nil, p.dbusPath, PinIFace)
	if err != nil {
		return
	}
	
	// introspection
	conn.Export(nil, p.dbusPath, introspect.IntrospectData.Name)
	if err != nil {
		return
	}
	
	return // OK
}

// is pin analog?
func (p *Pin) isAnalog() bool {
	return strings.HasPrefix(p.port, "AIN")
}

// initialize pin and allocate resources
// 'mode' can be "out", "in", "rising", "falling" or "both" for digital pins
// and polling period for analog pin in milliseconds
func (p *Pin) Attach(mode string) *dbus.Error {
	if p.isAnalog() {
//            self.m_adc_period = int(mode)
//            if self.m_adc_period > 0:
//                self.m_poll_flag = True
//                self.m_poll_thread = threading.Thread(target=self.analogPoller)
//                self.m_poll_thread.start()
	} else {
		if _, err := os.Stat(p.portDir); err == nil || !os.IsNotExist(err) { // TODO: check this condition! _.isDir()
			err := p.Detach()
			if err != nil {
				return err
			}
		}
		err := ioutil.WriteFile("/sys/class/gpio/export", []byte(p.port), 0200)
		if err != nil {
			return newDBusError(fmt.Sprintf("failed to export gpio: %s", err))
		}

		if mode == "out" {
			err := ioutil.WriteFile(fmt.Sprintf("%s/direction", p.portDir), []byte(mode), 0644)
			if err != nil {
				return newDBusError(fmt.Sprintf("failed to set OUT mode: %s", err))
			}
		} else {
			err := ioutil.WriteFile(fmt.Sprintf("%s/direction", p.portDir), []byte("in"), 0644)
			if err != nil {
				return newDBusError(fmt.Sprintf("failed to set IN mode: %s", err))
			}
			if mode != "in" {
				err := ioutil.WriteFile(fmt.Sprintf("%s/edge", p.portDir), []byte(mode), 0644)
				if err != nil {
					return newDBusError(fmt.Sprintf("failed to set edge: %s", err))
				}
//                    self.m_poll_flag = True
//                    self.m_poll_thread = threading.Thread(target=self.digitalPoller)
//                    self.m_poll_thread.start()
			} else {
				err := ioutil.WriteFile(fmt.Sprintf("%s/edge", p.portDir), []byte("none"), 0644)
				if err != nil {
					return newDBusError(fmt.Sprintf("failed to set edge: %s", err))
				}
			}
		}
	}

	p.attached = true
	return nil // OK
}

// Release pin and free all resources
func (p *Pin) Detach() *dbus.Error {
	err := ioutil.WriteFile("/sys/class/gpio/unexport", []byte(p.port), 0200)
	if err != nil {
		return newDBusError(fmt.Sprintf("failed to unexport gpio: %s", err))
	}
//        if self.m_poll_flag:
//            self.m_poll_flag = False
//            self.m_poll_thread.join()

	p.attached = false
	return nil // OK
}

// get pin value
func (p *Pin) GetValue() (string, *dbus.Error) {
	if !p.attached {
		return "", newDBusError(fmt.Sprintf("Pin is not attached"))
	}

	buf, err := ioutil.ReadFile(p.portVal)
	if err != nil {
		return "", newDBusError(fmt.Sprintf("Failed to read value: %s", err))
	}
	return strings.TrimSpace(string(buf)), nil
}

// set pin value
func (p *Pin) SetValue(val string) *dbus.Error {
	if !p.attached {
		return newDBusError(fmt.Sprintf("Pin is not attached"))
	}
	if p.isAnalog() {
		return newDBusError("This pin is input and cannot be set")
	}

	// truncate value?
	if val == "" || val == "0" {
		val = "0"
	} else {
		val = "1"
	}

	err := ioutil.WriteFile(p.portVal, []byte(val), 0644)
	if err != nil {
		return newDBusError(fmt.Sprintf("Failed to write value: %s", err))
	}

	return nil // OK
}

// set pin value to high
func (p *Pin) Set() *dbus.Error {
	return p.SetValue("1")
}

// set pin value to low
func (p *Pin) Clear() *dbus.Error {
	return p.SetValue("0")
}

// toggle pin value
func (p *Pin) Toggle() *dbus.Error {
	v, err := p.GetValue();
	if err != nil {
		return err
	}

	if v == "0" {
		return p.Set()
	} else {
		return p.Clear()
	}
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


// daemon entry point
func main() {
	bus, err := newDBus(BusName)
	if err != nil {
		log.Fatalf("no D-Bus access: %s", err)
	}

	s := newService()

	profile := flag.String("profile", "", "YML file or profile directory")
	flag.Parse()

	if len(profile) != 0 {
		info, err := os.Stat(profile)
		if err != nil {
			
		}
		
		if info.IsDir() {
			model, err := ioutil.ReadFile("/sys/firmware/devicetree/base/model")
			model = strings.TrimSpace(model)
//                with open(, 'r') as hwid:
//                    yamlpath = os.path.join(patharg, "{}.yaml".format(model))

//                    try:
//                        gpio_service.init_from_file(yamlpath)
//                    except FileExistsError:
//                        raise
//                    except IOError:
//                        raise FileNotFoundError("Profile file for {} not found.".format(model))                
		}
//        if os.path.isfile(patharg):
//            # loading pin profile from file
//            gpio_service.init_from_file(patharg)
//        elif os.path.isdir(patharg):
//            # loading ping profile from directory with profiles

//            try:

//            except IOError:
//                raise SystemError("Board not found.")
            
//        else:
//            print("Profile does not exist.")
//            return 1
	}

	err = s.export(bus)
	if err != nil {
		log.Fatalf("failed to export service: %s", err)
	}

	select {}
}
