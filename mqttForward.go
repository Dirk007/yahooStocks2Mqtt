package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/goiiot/libmqtt"
	log "github.com/sirupsen/logrus"
)

type Serializeable interface {
	Serialize() (string, error)
}

type MqttForwarder[V Serializeable] struct {
	config       MqttConfig
	publishTopic string
	commandTopic string
	killSwitch   chan bool
	data         chan V
}

type Credentials struct {
	Username string `yaml:""`
	Password string `yaml:""`
}

type MqttConfig struct {
	Host        string       `yaml:""`
	Port        uint16       `yaml:""`
	ClientID    *string      `yaml:""`
	Credentials *Credentials `yaml:""`
}

// ConnectionString returns the mqtt-client compatible connection-string from the config
func (config MqttConfig) ConnectionString() string {
	return fmt.Sprintf("%v:%v", config.Host, config.Port)
}

type MqttCommand struct {
	Command string `json:"command"`
}

// IsKill returns true if the given `command` represents a Kill
func (command MqttCommand) IsKill() bool {
	return strings.ToLower(command.Command) == "kill"
}

func (forwarder MqttForwarder[_]) onMqttConnected(client mqtt.Client, server string, code byte, err error) {
	logLine := fmt.Sprintf("Connection to MQTT. Server %v, Code %v, err %v", server, code, err)
	if err != nil || code != mqtt.CodeSuccess {
		log.Fatal(logLine)
	} else {
		log.Debug(logLine)
		client.Subscribe([]*mqtt.Topic{
			{Name: forwarder.commandTopic, Qos: mqtt.Qos1},
		}...)
	}
}

func (forwarder MqttForwarder[_]) onMqttMessage(client mqtt.Client, topic string, qos mqtt.QosLevel, msg []byte) {
	log.Debugf("MQTT [%v] message: %v", topic, string(msg))
	if topic != forwarder.commandTopic {
		return
	}

	command := MqttCommand{}
	err := json.Unmarshal(msg, &command)
	if err != nil {
		log.Warnf("Unable to unmarshal mqtt command %v: %v", string(msg), err)
		return
	}

	log.Debugf("Received mqtt command: '%v'", command)

	// TODO: Add more as needed
	if command.IsKill() {
		// TODO: Unsure - I think we need one write per listener.. so two at this time
		forwarder.killSwitch <- true
		forwarder.killSwitch <- true
	}

}

func (forwarder MqttForwarder[V]) Run() {

	options := []mqtt.Option{
		mqtt.WithKeepalive(60, 1.2),
		mqtt.WithAutoReconnect(true),
		mqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		mqtt.WithRouter(mqtt.NewRegexRouter()),
	}

	if forwarder.config.ClientID != nil {
		log.Debugf("Using MQTT clientID %v", forwarder.config.ClientID)
		options = append(options, mqtt.WithClientID(*forwarder.config.ClientID))
	}

	if forwarder.config.Credentials != nil {
		log.Debugf("Using MQTT credentials %v:*******", forwarder.config.Credentials.Username)
		options = append(options, mqtt.WithIdentity(forwarder.config.Credentials.Username, forwarder.config.Credentials.Password))
	}

	client, err := mqtt.NewClient(
		options...,
	)

	if err != nil {
		panic(fmt.Errorf("Unable to create MQTT client: %v", err))
	}

	client.ConnectServer(forwarder.config.ConnectionString(),
		mqtt.WithCustomTLS(nil),
		mqtt.WithConnHandleFunc(forwarder.onMqttConnected))

	// Disconnect nicely
	defer client.Destroy(false)

	client.HandleTopic(".*", forwarder.onMqttMessage)

	quote := *new(V)
	for {
		select {
		case quote = <-forwarder.data:
			log.Debugf("Received quote: %v", quote)
			jsonQuote, err := quote.Serialize()
			if err != nil {
				log.Warnf("Error marshalling incoming quote: %v", err)
				continue
			}
			client.Publish([]*mqtt.PublishPacket{
				{
					TopicName: forwarder.publishTopic,
					Payload:   []byte(jsonQuote),
					Qos:       mqtt.Qos1,
				},
			}...)
		case _ = <-forwarder.killSwitch:
			log.Info("Exiting mqtt for killswitch")
			break
		}
	}
}
