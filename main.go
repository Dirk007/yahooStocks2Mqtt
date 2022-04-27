package main

import (
	"os"

	mqtt "github.com/Dirk007/yahooQuotes/pkg/mqtt"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	configFile := os.Getenv("CONFIG_FILE")
	if len(configFile) == 0 {
		configFile = "config.yaml"
		log.Debugf("Using config file %s", configFile)
	}
	configuration, err := GetConfig(configFile); if err != nil {
		configuration = GetDefaultConfig()
	}

	quotes := make(chan YahooStockInfo)
	kill := make(chan bool)

	forwarder, err := mqtt.NewForwarderBuilder[YahooStockInfo]().
		Channels().Data(quotes).KillWitch(kill).
		Topics().Input("stocks/command").Output("stocks/quote").
		Server().Host(configuration.Mqtt.Host).Port(configuration.Mqtt.Port).
		MQTT().Config(configuration.Mqtt).
		Build()

	if err != nil {
		log.Errorf("Unable to build MQTT forwarder: %v", err)
		os.Exit(1)
	}

	go requestLoop(configuration.Symbols, configuration.RequestPeriod(), quotes, kill)

	forwarder.Run()
}
