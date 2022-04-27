package main

import (
	"io/ioutil"
	"time"

	mqtt "github.com/Dirk007/yahooQuotes/pkg/mqtt"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Config represents the configuration of the application
type Config struct {
	RequestPeriodRepr string     `yaml:"requestperiod"` //< humantime
	Symbols           []string   `yamml:""`
	Mqtt              mqtt.Config `yaml:""`
}

// RequestPeriod returns the Duration of the human readable `RequestPeriodRepr` field or
// a default value if that field in the config is malformed.
func (config Config) RequestPeriod() time.Duration {
	const DefaultPeriod = 5 * time.Minute

	requestPeriod, timeError := time.ParseDuration(config.RequestPeriodRepr)
	if timeError != nil {
		log.Error("Warning: unable to parse RequestPeriod. Using internal default. Error: %v", timeError)
		requestPeriod = DefaultPeriod
	}

	return requestPeriod
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	return &Config{
		RequestPeriodRepr: "5m",
		Symbols:           []string{"VGWL.DE", "VFEM.DE", "NLLSF", "PFE.DE"},
		Mqtt: mqtt.Config{
			Host: "192.168.1.104",
			Port: mqtt.DefaultMqttPort,
		},
	}
}

// GetConfig reads the config from the given filename
func GetConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Errorf("Unable to read config from %v, using defauklts. Error: %v\n", filename, err)
		return nil, err
	}

	config := Config{}
	if yaml.Unmarshal(content, &config) != nil {
		log.Errorf("Unable to unmarshal config content '%v'. Error: %v\n", content, err)
		return nil, err
	}

	return &config, nil
}
