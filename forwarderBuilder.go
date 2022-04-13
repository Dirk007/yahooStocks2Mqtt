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

func (builder *ForwarderBuilder[V]) Build() (*MqttForwarder[V], error) {
	if builder.inner.data == nil ||
		builder.inner.killSwitch == nil ||
		builder.inner.publishTopic == "" ||
		builder.inner.commandTopic == "" ||
		builder.inner.config.Host == "" ||
		builder.inner.config.Port == 0 {
		return nil, fmt.Errorf("Builder not ready - missing settings")
	}

	return builder.inner, nil
}
