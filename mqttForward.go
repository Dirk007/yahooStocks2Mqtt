package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/goiiot/libmqtt"
)

const STOCK_COMMAND_TOPIC = "stock/command"
const STOCK_PUBLISH_TOPIC = "stock/quote"

type MqttConfig struct {
	MqttHost string `yaml:""`
	MqttPort uint16 `yaml:""`
}

// ConnectionString returns the mqtt-client compatible connection-string from the config
func (config MqttConfig) ConnectionString() string {
	return fmt.Sprintf("%v:%v", config.MqttHost, config.MqttPort)
}

type MqttCommand struct {
	Command string `json:"command"`
}

// IsKill returns true if the given `command` represents a Kill
func (command MqttCommand) IsKill() bool {
	return strings.ToLower(command.Command) == "kill"
}

func onMqttConnected(client mqtt.Client, server string, code byte, err error) {
	logLine := fmt.Sprintf("Connection to MQTT. Server %v, Code %v, err %v", server, code, err)
	if err != nil || code != mqtt.CodeSuccess {
		log.Fatal(logLine)
	} else {
		log.Println(logLine)
		client.Subscribe([]*mqtt.Topic{
			{Name: STOCK_COMMAND_TOPIC, Qos: mqtt.Qos1},
		}...)
	}
}

type Serializeable interface {
	Serialize() (string, error)
}

func mqttLoop[V Serializeable](config MqttConfig, quotesChannel chan V, kill chan bool) {
	client, err := mqtt.NewClient(
		mqtt.WithKeepalive(60, 1.2),
		mqtt.WithAutoReconnect(true),
		mqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		mqtt.WithRouter(mqtt.NewRegexRouter()),
	)

	if err != nil {
		panic(fmt.Errorf("Unable to create MQTT client: %v", err))
	}

	client.ConnectServer(config.ConnectionString(),
		mqtt.WithCustomTLS(nil),
		mqtt.WithConnHandleFunc(onMqttConnected))

	// Disconnect nicely
	defer client.Destroy(false)

	client.HandleTopic(".*", func(client mqtt.Client, topic string, qos mqtt.QosLevel, msg []byte) {
		log.Printf("MQTT [%v] message: %v", topic, string(msg))
		if topic != STOCK_COMMAND_TOPIC {
			return
		}

		command := MqttCommand{}
		err := json.Unmarshal(msg, &command)
		if err != nil {
			log.Printf("Unable to unmarshal mqtt command %v: %v", string(msg), err)
			return
		}

		log.Printf("Received mqtt command: '%v'", command)

		// TODO: Add more as needed
		if command.IsKill() {
			kill <- true
		}
	})

	for {
		quote := <-quotesChannel
		log.Printf("Received quote: %v", quote)
		jsonQuote, err := quote.Serialize()
		if err != nil {
			log.Printf("Error marshalling incoming quote: %v", err)
			continue
		}
		client.Publish([]*mqtt.PublishPacket{
			{
				TopicName: STOCK_PUBLISH_TOPIC,
				Payload:   []byte(jsonQuote),
				Qos:       mqtt.Qos0,
			},
		}...)
	}
}
