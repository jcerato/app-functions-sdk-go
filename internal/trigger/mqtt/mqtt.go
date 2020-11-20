//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package mqtt

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/google/uuid"

	"github.com/jcerato/app-functions-sdk-go/appcontext"
	"github.com/jcerato/app-functions-sdk-go/internal/common"
	"github.com/jcerato/app-functions-sdk-go/internal/runtime"
	"github.com/jcerato/app-functions-sdk-go/internal/security"
	"github.com/jcerato/app-functions-sdk-go/pkg/secure"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	configuration  *common.ConfigurationStruct
	mqttClient     pahoMqtt.Client
	runtime        *runtime.GolangRuntime
	edgeXClients   common.EdgeXClients
	secretProvider security.SecretProvider
}

func NewTrigger(
	configuration *common.ConfigurationStruct,
	runtime *runtime.GolangRuntime,
	clients common.EdgeXClients,
	secretProvider security.SecretProvider) *Trigger {
	return &Trigger{
		configuration:  configuration,
		runtime:        runtime,
		edgeXClients:   clients,
		secretProvider: secretProvider,
	}
}

// Initialize initializes the Trigger for an external MQTT broker
func (trigger *Trigger) Initialize(_ *sync.WaitGroup, _ context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	// Convenience short cuts
	logger := trigger.edgeXClients.LoggingClient
	brokerConfig := trigger.configuration.MqttBroker
	topic := trigger.configuration.Binding.SubscribeTopic

	logger.Info("Initializing MQTT Trigger")

	if background != nil {
		return nil, errors.New("background publishing not supported for services using MQTT trigger")
	}

	if len(topic) == 0 {
		return nil, fmt.Errorf("missing SubscribeTopic for MQTT Trigger. Must be present in [Binding] section.")
	}

	brokerUrl, err := url.Parse(brokerConfig.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT Broker Url '%s': %s", trigger.configuration.MqttBroker.Url, err.Error())
	}

	opts := pahoMqtt.NewClientOptions()
	opts.AutoReconnect = brokerConfig.AutoReconnect
	opts.OnConnect = trigger.onConnectHandler
	opts.ClientID = brokerConfig.ClientId
	if len(brokerConfig.ConnectTimeout) > 0 {
		duration, err := time.ParseDuration(brokerConfig.ConnectTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid MQTT ConnectTimeout '%s': %s", brokerConfig.ConnectTimeout, err.Error())
		}
		opts.ConnectTimeout = duration
	}
	opts.KeepAlive = brokerConfig.KeepAlive
	opts.Servers = []*url.URL{brokerUrl}

	mqttFactory := secure.NewMqttFactory(
		logger,
		trigger.secretProvider,
		brokerConfig.AuthMode,
		brokerConfig.SecretPath,
		brokerConfig.SkipCertVerify,
	)

	mqttClient, err := mqttFactory.Create(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to create secure MQTT Client: %s", err.Error())
	}

	logger.Info(fmt.Sprintf("Connecting to mqtt broker for MQTT trigger at: %s", brokerUrl))

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to broker for MQTT trigger: %s", token.Error().Error())
	}

	logger.Info("Connected to mqtt server for MQTT trigger")

	deferred := func() {
		logger.Info("Disconnecting from broker for MQTT trigger")
		trigger.mqttClient.Disconnect(0)
	}

	trigger.mqttClient = mqttClient

	return deferred, nil
}

func (trigger *Trigger) onConnectHandler(mqttClient pahoMqtt.Client) {
	// Convenience short cuts
	logger := trigger.edgeXClients.LoggingClient
	topic := trigger.configuration.Binding.SubscribeTopic
	qos := trigger.configuration.MqttBroker.QoS

	if token := mqttClient.Subscribe(topic, qos, trigger.messageHandler); token.Wait() && token.Error() != nil {
		mqttClient.Disconnect(0)
		logger.Error(fmt.Sprintf("could not subscribe to topic '%s' for MQTT trigger: %s",
			topic, token.Error().Error()))
		return
	}

	logger.Info(fmt.Sprintf("Subscribed to topic '%s' for MQTT trigger", topic))
}

func (trigger *Trigger) messageHandler(client pahoMqtt.Client, message pahoMqtt.Message) {
	// Convenience short cuts
	logger := trigger.edgeXClients.LoggingClient
	brokerConfig := trigger.configuration.MqttBroker
	topic := trigger.configuration.Binding.PublishTopic

	data := message.Payload()
	contentType := clients.ContentTypeJSON
	if data[0] != byte('{') {
		// If not JSON then assume it is CBOR
		contentType = clients.ContentTypeCBOR
	}

	correlationID := uuid.New().String()

	edgexContext := &appcontext.Context{
		CorrelationID:         correlationID,
		Configuration:         trigger.configuration,
		LoggingClient:         trigger.edgeXClients.LoggingClient,
		EventClient:           trigger.edgeXClients.EventClient,
		ValueDescriptorClient: trigger.edgeXClients.ValueDescriptorClient,
		CommandClient:         trigger.edgeXClients.CommandClient,
		NotificationsClient:   trigger.edgeXClients.NotificationsClient,
	}

	logger.Trace("Received message from MQTT Trigger", clients.CorrelationHeader, correlationID)
	logger.Debug(fmt.Sprintf("Received message from MQTT Trigger with %d bytes", len(data)), clients.ContentType, contentType)

	envelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
	}

	messageError := trigger.runtime.ProcessMessage(edgexContext, envelope)
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		// ToDo: Do we want to publish the error back to the Broker?
		return
	}

	if len(edgexContext.OutputData) > 0 && len(topic) > 0 {
		if token := client.Publish(topic, brokerConfig.QoS, brokerConfig.Retain, edgexContext.OutputData); token.Wait() && token.Error() != nil {
			logger.Error("could not publish to topic '%s' for MQTT trigger: %s", topic, token.Error().Error())
		} else {
			logger.Trace("Sent MQTT Trigger response message", clients.CorrelationHeader, correlationID)
			logger.Debug(fmt.Sprintf("Sent MQTT Trigger response message on topic '%s' with %d bytes", topic, len(edgexContext.OutputData)))
		}
	}
}
