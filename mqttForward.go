package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/goiiot/libmqtt"
)

type Serializeable interface {
	Serialize() (string, error)
}

type MqttForwarder[V Serializeable] struct {
	target       string
	publishTopic string
	commandTopic string
	killSwitch   chan bool
	data         chan V
}

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

func NewForwarder[V Serializeable](config MqttConfig, publishTopic string, commandTopic string, killSwitch chan bool, data chan V) MqttForwarder[V] {
	return MqttForwarder[V]{
		target:       config.ConnectionString(),
		publishTopic: publishTopic,
		commandTopic: commandTopic,
		killSwitch:   killSwitch,
		data:         data,
	}
}

func (forwarder MqttForwarder[_]) onMqttConnected(client mqtt.Client, server string, code byte, err error) {
	logLine := fmt.Sprintf("Connection to MQTT. Server %v, Code %v, err %v", server, code, err)
	if err != nil || code != mqtt.CodeSuccess {
		log.Fatal(logLine)
	} else {
		log.Println(logLine)
		client.Subscribe([]*mqtt.Topic{
			{Name: forwarder.commandTopic, Qos: mqtt.Qos1},
		}...)
	}
}

func (forwarder MqttForwarder[_]) onMqttMessage(client mqtt.Client, topic string, qos mqtt.QosLevel, msg []byte) {
	log.Printf("MQTT [%v] message: %v", topic, string(msg))
	if topic != forwarder.commandTopic {
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
		forwarder.killSwitch <- true
	}

}

func (forwarder MqttForwarder[V]) Run() {
	client, err := mqtt.NewClient(
		mqtt.WithKeepalive(60, 1.2),
		mqtt.WithAutoReconnect(true),
		mqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		mqtt.WithRouter(mqtt.NewRegexRouter()),
	)

	if err != nil {
		panic(fmt.Errorf("Unable to create MQTT client: %v", err))
	}

	client.ConnectServer(forwarder.target,
		mqtt.WithCustomTLS(nil),
		mqtt.WithConnHandleFunc(forwarder.onMqttConnected))

	// Disconnect nicely
	defer client.Destroy(false)

	client.HandleTopic(".*", forwarder.onMqttMessage)

	quote := *new(V)
	for {
		select {
		case quote = <-forwarder.data:
			log.Printf("Received quote: %v", quote)
			jsonQuote, err := quote.Serialize()
			if err != nil {
				log.Printf("Error marshalling incoming quote: %v", err)
				continue
			}
			client.Publish([]*mqtt.PublishPacket{
				{
					TopicName: forwarder.publishTopic,
					Payload:   []byte(jsonQuote),
					Qos:       mqtt.Qos0,
				},
			}...)
		case _ = <-forwarder.killSwitch:
			break
		}
	}
}
