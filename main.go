package main

import (
	"os"
)

func main() {
	config_file := os.Getenv("CONFIG_FILE")
	if config_file == "" {
		config_file = "config.yaml"
	}
	config := getConfig(config_file)

	quotes := make(chan YahooStockInfo)
	kill := make(chan bool)

	go requestLoop(config.Symbols, config.RequestPeriod(), quotes, kill)

	forwarder := NewForwarder(config.Mqtt, "stocks/quote", "stocks/command", kill, quotes)
	forwarder.Run()
}
