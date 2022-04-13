package main

import "fmt"

const DEFAULT_MQTT_PORT = 1883

type ForwarderBuilder[V Serializeable] struct {
	inner *MqttForwarder[V]
}

type ForwarderTopicBuilder[V Serializeable] struct {
	ForwarderBuilder[V]
}

type ForwarderChannelBuilder[V Serializeable] struct {
	ForwarderBuilder[V]
}

type ForwarderConnectionBuilder[V Serializeable] struct {
	ForwarderBuilder[V]
}

type ForwarderMQTTBuilder[V Serializeable] struct {
	ForwarderBuilder[V]
}

type ForwarderCredentialsBuilder[V Serializeable] struct {
	ForwarderMQTTBuilder[V]
}

func NewForwarderBuilder[V Serializeable]() *ForwarderBuilder[V] {
	return &ForwarderBuilder[V]{
		inner: &MqttForwarder[V]{},
	}
}

func (builder *ForwarderBuilder[V]) Topics() *ForwarderTopicBuilder[V] {
	return &ForwarderTopicBuilder[V]{*builder}
}

func (builder *ForwarderTopicBuilder[V]) Output(topic string) *ForwarderTopicBuilder[V] {
	builder.inner.publishTopic = topic
	return builder
}

func (builder *ForwarderTopicBuilder[V]) Input(topic string) *ForwarderTopicBuilder[V] {
	builder.inner.commandTopic = topic
	return builder
}

func (builder *ForwarderBuilder[V]) Channels() *ForwarderChannelBuilder[V] {
	return &ForwarderChannelBuilder[V]{*builder}
}

func (builder *ForwarderChannelBuilder[V]) Data(data chan V) *ForwarderChannelBuilder[V] {
	builder.inner.data = data
	return builder
}

func (builder *ForwarderChannelBuilder[V]) KillWitch(killSwitch chan bool) *ForwarderChannelBuilder[V] {
	builder.inner.killSwitch = killSwitch
	return builder
}

func (builder *ForwarderBuilder[V]) Server() *ForwarderConnectionBuilder[V] {
	return &ForwarderConnectionBuilder[V]{*builder}
}

func (builder *ForwarderConnectionBuilder[V]) Host(host string) *ForwarderConnectionBuilder[V] {
	builder.inner.config.Host = host
	return builder
}

func (builder *ForwarderConnectionBuilder[V]) Port(port uint16) *ForwarderConnectionBuilder[V] {
	builder.inner.config.Port = port
	return builder
}

func (builder *ForwarderConnectionBuilder[V]) DefaultPort() *ForwarderConnectionBuilder[V] {
	builder.inner.config.Port = DEFAULT_MQTT_PORT
	return builder
}

func (builder *ForwarderBuilder[V]) MQTT() *ForwarderMQTTBuilder[V] {
	return &ForwarderMQTTBuilder[V]{*builder}
}

func (builder *ForwarderMQTTBuilder[V]) Config(config MqttConfig) *ForwarderMQTTBuilder[V] {
	builder.inner.config = config
	return builder
}

func (builder *ForwarderMQTTBuilder[V]) ClientID(clientID string) *ForwarderMQTTBuilder[V] {
	builder.inner.config.ClientID = &clientID
	return builder
}

func (builder *ForwarderMQTTBuilder[V]) Credentials() *ForwarderCredentialsBuilder[V] {
	result := &ForwarderCredentialsBuilder[V]{*builder}
	result.inner.config.Credentials = &Credentials{}
	return result
}

func (builder *ForwarderCredentialsBuilder[V]) Username(username string) *ForwarderCredentialsBuilder[V] {
	builder.inner.config.Credentials.Username = username
	return builder
}

func (builder *ForwarderCredentialsBuilder[V]) Password(password string) *ForwarderCredentialsBuilder[V] {
	builder.inner.config.Credentials.Password = password
	return builder
}

func credentialsOK(config MqttConfig) bool {
	return (config.Credentials != nil &&
		config.Credentials.Username != "" &&
		config.Credentials.Password != "") ||
		config.Credentials == nil
}

func (builder *ForwarderBuilder[V]) Build() (*MqttForwarder[V], error) {
	if builder.inner.data == nil ||
		builder.inner.killSwitch == nil ||
		builder.inner.publishTopic == "" ||
		builder.inner.commandTopic == "" ||
		builder.inner.config.Host == "" ||
		builder.inner.config.Port == 0 ||
		!credentialsOK(builder.inner.config) {
		return nil, fmt.Errorf("Builder not ready - missing settings")
	}

	return builder.inner, nil
}
