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

	builder := NewForwarderBuilder[YahooStockInfo]().
		Channels().Data(quotes).KillWitch(kill).
		Topics().Input("stocks/command").Output("stockts/quote").
		Server().Host(config.Mqtt.Host).Port(config.Mqtt.Port)

	if config.Mqtt.ClientID != nil {
		builder.MQTT().ClientID(*config.Mqtt.ClientID)
	}

	if config.Mqtt.Credentials != nil {
		builder.MQTT().Credentials().
			Username(config.Mqtt.Credentials.Username).
			Password(config.Mqtt.Credentials.Password)
	}

	forwarder, err := builder.Build()

	if err != nil {
		log.Errorf("Unable to build MQTT forwarder: %v", err)
		os.Exit(1)
	}

	go requestLoop(config.Symbols, config.RequestPeriod(), quotes, kill)

	forwarder.Run()
}
