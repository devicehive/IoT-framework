package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	// service configuration
	config *Config
)

// initialize command line options and parse them.
func init() {
	cfgFileName := flag.String("conf", "", "configuration file (YAML format)")

	if !flag.Parsed() {
		flag.Parse()
	}

	if cfgFileName == nil || len(*cfgFileName) == 0 {
		// no file provided
		flag.Usage()
		os.Exit(1)
	}

	config = new(Config)
	if err := config.FromFile(*cfgFileName); err != nil {
		panic(err) // failed to parse configuration
	}
}

// Config holds all configuration data.
type Config struct {
	URL       string `yaml:"URL,omitempty"`
	AccessKey string `yaml:"AccessKey,omitempty"`

	DeviceID   string `yaml:"DeviceID,omitempty"`
	DeviceName string `yaml:"DeviceName,omitempty"`
	DeviceKey  string `yaml:"DeviceKey,omitempty"`

	NetworkName string `yaml:"NetworkName,omitempty"`
	NetworkKey  string `yaml:"NetworkKey,omitempty"`
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
