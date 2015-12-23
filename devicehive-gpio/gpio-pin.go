package main // TODO: move to special package?

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	PinPathBase = "/com/devicehive/gpio"
	PinIFace    = "com.devicehive.gpio.Pin"
)

// exported pin
type Pin struct {
	name string
	port string

	dbusConn *dbus.Conn
	dbusPath dbus.ObjectPath

	portDir string // port directory path
	portVal string // value file

	pollPeriod time.Duration // for analog input
	stopChan   chan bool

	started bool
}

// create new pin
func NewPin(pin, port string) *Pin {
	p := &Pin{}
	p.name = pin
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

	p.dbusConn = conn
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
func (p *Pin) start(mode string) (needPoll bool, err error) {
	// analog pin
	if p.isAnalog() {
		p.pollPeriod, err = time.ParseDuration(mode)
		if err != nil {
			return
		}
		return true, nil
	}

	// stop digital pin (if exists)
	if info, err := os.Stat(p.portDir); err == nil && info.IsDir() {
		err = p.stop()
		if err != nil {
			return false, err
		}
	}

	// export
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte(p.port), 0200)
	if err != nil {
		return false, fmt.Errorf("failed to export: %s", err)
	}

	if mode == "out" {
		// set OUT mode
		err = ioutil.WriteFile(fmt.Sprintf("%s/direction", p.portDir), []byte(mode), 0644)
		if err != nil {
			return false, fmt.Errorf("failed to set OUT mode: %s", err)
		}
		return false, nil // OK
	} else {
		// set IN mode
		err = ioutil.WriteFile(fmt.Sprintf("%s/direction", p.portDir), []byte("in"), 0644)
		if err != nil {
			return false, fmt.Errorf("failed to set IN mode: %s", err)
		}

		if mode != "in" {
			// set EDGE
			err = ioutil.WriteFile(fmt.Sprintf("%s/edge", p.portDir), []byte(mode), 0644)
			if err != nil {
				return false, fmt.Errorf("failed to set edge: %s", err)
			}
			return true, nil // OK
		} else {
			err = ioutil.WriteFile(fmt.Sprintf("%s/edge", p.portDir), []byte("none"), 0644)
			if err != nil {
				return false, fmt.Errorf("failed to set edge: %s", err)
			}
		}

		return false, nil // OK
	}
}

// initialize pin and allocate resources
// 'mode' can be "out", "in", "rising", "falling" or "both" for digital pins
// and polling period for analog pin in milliseconds
func (p *Pin) Start(mode string) *dbus.Error {
	needPoll, err := p.start(mode)
	if err != nil {
		return newDBusError(err.Error())
	}

	// start polling
	if needPoll {
		p.stopChan = make(chan bool, 1)
		if p.isAnalog() {
			go p.analogPoll()
		} else {
			go p.digitalPoll()
		}
	}

	p.started = true
	return nil // OK
}

// Release pin and free all resources
func (p *Pin) stop() (err error) {
	err = ioutil.WriteFile("/sys/class/gpio/unexport", []byte(p.port), 0200)
	if err != nil {
		return fmt.Errorf("failed to unexport gpio: %s", err)
	}

	return nil // OK
}

// Release pin and free all resources
func (p *Pin) Stop() *dbus.Error {
	// stop polling if any
	if p.stopChan != nil {
		p.stopChan <- true
	}

	err := p.stop()
	if err != nil {
		return newDBusError(err.Error())
	}

	p.started = false
	return nil // OK
}

// emit ValueChanged signal
func (p *Pin) emitValueChanged(val string) (err error) {
	if p.dbusConn != nil {
		err = p.dbusConn.Emit(p.dbusPath,
			fmt.Sprintf("%s.ValueChanged", PinIFace),
			p.name, val)
	} else {
		log.Printf("WARN: pin:%q is not exported on D-Bus to emit signals", p.name)
	}

	return
}

// analog pin poller
func (p *Pin) analogPoll() {
	for {
		select {
		case <-time.After(p.pollPeriod):
			val, err := p.Get()
			if err != nil {
				log.Printf("WARN: pin:%q failed to read value: %s", p.name, err)
			} else {
				p.emitValueChanged(val)
			}
		case <-p.stopChan:
			log.Printf("analog polling stopped pin:%q", p.name)
			return
		}
	}
}

// digital pin poller
func (p *Pin) digitalPoll() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("WARN: pin:%q failed to start watcher: %s", p.name, err)
		return
	}
	defer w.Close()

	err = w.Add(p.portVal)
	if err != nil {
		log.Printf("WARN: pin:%q failed to watch pin value: %s", p.name, err)
		return
	}

	oldval, _ := p.GetValue()
	for {
		select {
		case e := <-w.Events:
			// log.Printf("DEBUG: watcher event: %v", e)
			if (e.Op & fsnotify.Write) != 0 {
				val, _ := p.Get()
				if val != oldval {
					p.emitValueChanged(val)
					oldval = val
				}
			}

		case err := <-w.Errors:
			log.Printf("WARN: watcher error: %s", err)

		case <-p.stopChan:
			log.Printf("digital polling stopped pin:%q", p.name)
			return
		}
	}
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

	p.emitValueChanged(val)
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
