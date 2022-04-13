package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}
	config := getConfig(configFile)

	quotes := make(chan YahooStockInfo)
	kill := make(chan bool)

	forwarder, err := NewForwarderBuilder[YahooStockInfo]().
		Channels().Data(quotes).KillWitch(kill).
		Topics().Input("stocks/command").Output("stockts/quote").
		Server().Host(config.Mqtt.Host).Port(config.Mqtt.Port).
		MQTT().Config(config.Mqtt).
		Build()

	if err != nil {
		log.Errorf("Unable to build MQTT forwarder: %v", err)
		os.Exit(1)
	}

	go requestLoop(config.Symbols, config.RequestPeriod(), quotes, kill)

	forwarder.Run()
}
