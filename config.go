package main

import (
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	RequestPeriodRepr string     `yaml:"requestperiod"` //< humantime
	Symbols           []string   `yamml:""`
	Mqtt              MqttConfig `yaml:""`
}

// RequestPeriod returns the Duration of the human readable `RequestPeriodRepr` field or
// a default value if that field in the config is malformed.
func (config Config) RequestPeriod() time.Duration {
	const DEFAULT_PERIOD = 5 * time.Minute

	requestPeriod, timeError := time.ParseDuration(config.RequestPeriodRepr)
	if timeError != nil {
		log.Error("Warning: unable to parse RequestPeriod. Using internal default. Error: %v", timeError)
		requestPeriod = DEFAULT_PERIOD
	}

	return requestPeriod
}

func getDefaultConfig() Config {
	return Config{
		RequestPeriodRepr: "5m",
		Symbols:           []string{"VGWL.DE", "VFEM.DE", "NLLSF", "PFE.DE"},
		Mqtt: MqttConfig{
			Host: "192.168.1.104",
			Port: 1883,
		},
	}
}

func getConfig(from string) Config {
	content, err := ioutil.ReadFile(from)

	if err != nil {
		log.Error("Unable to read config from %v, using defauklts. Error: %v\n", from, err)
		return getDefaultConfig()
	}

	config := Config{}
	if yaml.Unmarshal(content, &config) != nil {
		log.Warn("Unable to unmarshal config content '%v'. Error: %v\n", content, err)
		return getDefaultConfig()
	}

	return config
}
