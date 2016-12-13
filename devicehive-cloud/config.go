package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	configFileName string
)

// initialize command line options and parse them.
func init() {
	flag.StringVar(&configFileName, "conf", "", "configuration file (YAML format)")
}

// Config holds all configuration data.
type Config struct {
	URL        string `yaml:"URL,omitempty"`
	RefreshKey string `yaml:"RefreshKey,omitempty"`
	AccessKey  string `yaml:"AccessKey,omitempty"`

	DeviceID   string `yaml:"DeviceID,omitempty"`
	DeviceName string `yaml:"DeviceName,omitempty"`

	NetworkName string `yaml:"NetworkName,omitempty"`
	NetworkDesc string `yaml:"NetworkDescription,omitempty"`

	LoggingLevel       string `yaml:"LoggingLevel,omitempty"`
	DeviceHiveLogLevel string `yaml:"DH-LoggingLevel,omitempty"`
}

// FromFile reads configuration from file.
func (cfg *Config) FromFile(filePath string) error {
	// read full configuration file
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read configuration from %q: %s", filePath, err)
	}

	// decode configuration
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return fmt.Errorf("failed to parse configuration from %q: %s", filePath, err)
	}

	return nil // OK
}
