package main // TODO: move to special package?

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"io/ioutil"
	"os"
	"strings"
)

const (
	PinPathBase = "/com/devicehive/gpio"
	PinIFace    = "com.devicehive.gpio.Pin"
)

// exported pin
type Pin struct {
	pin      string
	port     string
	dbusPath dbus.ObjectPath

	portDir string // port directory path
	portVal string // value file

	started bool
}

// create new pin
func NewPin(pin, port string) *Pin {
	p := &Pin{}
	p.pin = pin
	p.port = port
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
	conn.Export(nil, p.dbusPath,
		introspect.IntrospectData.Name)
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
func (p *Pin) Start(mode string) *dbus.Error {
	if p.isAnalog() {
		//            self.m_adc_period = int(mode)
		//            if self.m_adc_period > 0:
		//                self.m_poll_flag = True
		//                self.m_poll_thread = threading.Thread(target=self.analogPoller)
		//                self.m_poll_thread.start()
	} else {
		if _, err := os.Stat(p.portDir); err == nil || !os.IsNotExist(err) { // TODO: check this condition! _.isDir()
			err := p.Stop()
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

	p.started = true
	return nil // OK
}

// Release pin and free all resources
func (p *Pin) Stop() *dbus.Error {
	err := ioutil.WriteFile("/sys/class/gpio/unexport", []byte(p.port), 0200)
	if err != nil {
		return newDBusError(fmt.Sprintf("failed to unexport gpio: %s", err))
	}
	//        if self.m_poll_flag:
	//            self.m_poll_flag = False
	//            self.m_poll_thread.join()

	p.started = false
	return nil // OK
}

// get pin value
func (p *Pin) GetValue() (string, *dbus.Error) {
	if !p.started {
		return "", newDBusError(fmt.Sprintf("Pin is not started"))
	}

	buf, err := ioutil.ReadFile(p.portVal)
	if err != nil {
		return "", newDBusError(fmt.Sprintf("Failed to read pin value: %s", err))
	}
	return strings.TrimSpace(string(buf)), nil
}

// set pin value
func (p *Pin) SetValue(val string) *dbus.Error {
	if !p.started {
		return newDBusError(fmt.Sprintf("Pin is not started"))
	}
	if p.isAnalog() {
		return newDBusError("This pin is input and cannot be set")
	}

	// truncate value
	// FIXME: do we need this truncation?
	if val == "" || val == "0" {
		val = "0"
	} else {
		val = "1"
	}

	err := ioutil.WriteFile(p.portVal, []byte(val), 0644)
	if err != nil {
		return newDBusError(fmt.Sprintf("Failed to write pin value: %s", err))
	}

	return nil // OK
}

// get pin value (alias)
func (p *Pin) Get() (string, *dbus.Error) {
	return p.GetValue()
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
	v, err := p.GetValue()
	if err != nil {
		return err
	}

	if v == "0" {
		return p.Set()
	} else {
		return p.Clear()
	}
}
