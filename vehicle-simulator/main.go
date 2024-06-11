// Copyright (c) 2023 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-IDentifier: EPL-2.0 OR Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	dittoModel "github.com/eclipse/ditto-clients-golang/model"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/eclipse/ditto-clients-golang/protocol/things"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type deviceData struct {
	DeviceID string `json:"deviceID"`
	TenantID string `json:"tenantID"`
	PolicyID string `json:"policyID"`
}

var mqttClient mqtt.Client
var currentDeviceData = deviceData{}
var telemetrySenderStarted bool
var ctdSubscriptionTopics []string

var stopTelemetrySender = make(chan struct{})
var controlTelemetrySender = make(chan string)

// Define flag variables
var ctdEffectDelay = flag.Int("ctd-effect-delay", 30, "In seconds, Configure trunk/alarms state time")
var periodicUpdatesDelay = flag.Int("periodic-updates-delay", 5, "In seconds, Configure periodic updates (minimum 5)")
var idleStateTime = flag.Int("idle-state-time", 120, "In seconds, Configure when the system will enter idle state. IDle state will send telemetry data infrequently. Receiving CTD command will reset the idle state.")
var brokerURL = flag.String("broker-url", "tcp://localhost:1883", "Specify the MQTT broker URL to connect to (default \"tcp://localhost:1883\").")
var brokerUsername = flag.String("broker-username", "", "Specify the MQTT client password to authenticate with.")
var brokerPassword = flag.String("broker-password", "", "Specify the MQTT client username to authenticate with.")

var initialPeriodicUpdatesDelay = 0 // set the initial periodic updates delay in case of idle state active

func main() {
	flag.Parse()
	validateFlags()

	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	initializeClient()
	getDeviceData()

	defer func() {
		if telemetrySenderStarted {
			stopTelemetrySender <- struct{}{}
		}
	}()
	defer mqttClient.Disconnect(1000)

	<-keepAlive
}

func initializeClient() {
	slog.Info("Broker configuration", "URL", *brokerURL)
	options := mqtt.NewClientOptions()
	options.AddBroker(*brokerURL)
	if *brokerUsername != "" {
		options.SetUsername(*brokerUsername)
		options.SetPassword(*brokerPassword)
	}
	mqttClient = mqtt.NewClient(options)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func getEnvVariable(variableName string, defaultValue string) string {
	value := os.Getenv(variableName)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getDeviceData() {
	mqttClient.Subscribe("edge/thing/response", 1, handleDeviceDataReceived).Wait()
	mqttClient.Publish("edge/thing/request", 1, false, "{}").Wait()
}

func handleDeviceDataReceived(sourceClient mqtt.Client, message mqtt.Message) {
	deviceDataReceived := deviceData{}
	if err := json.Unmarshal(message.Payload(), &deviceDataReceived); err != nil {
		slog.Error("Error during payload unmarshal: ", err)
		return
	}

	slog.Info("Connector Device Data received successfully", "data", deviceDataReceived)
	if deviceDataReceived.DeviceID == currentDeviceData.DeviceID && deviceDataReceived.TenantID == currentDeviceData.TenantID {
		slog.Debug("Skipping the same connector instance.")
		return
	}

	unsubscribeAllCTD()
	subscribeForCTD(deviceDataReceived.DeviceID)
	currentDeviceData = deviceDataReceived
	startSendingRandomlyGeneratedTelemetryData()
}

func subscribeForCTD(deviceID string) {
	slog.Info("Subscribing for CTD commands: resetDTC & openTrunk...")
	topicResetDTC := fmt.Sprintf("command//%s:Vehicle/req//resetDTC", deviceID)
	ctdSubscriptionTopics = append(ctdSubscriptionTopics, topicResetDTC)
	topicOpenTrunk := fmt.Sprintf("command//%s:Vehicle/req//openTrunk", deviceID)
	ctdSubscriptionTopics = append(ctdSubscriptionTopics, topicOpenTrunk)

	mqttClient.Subscribe(topicResetDTC, 1, handleResetDTC).Wait()
	mqttClient.Subscribe(topicOpenTrunk, 1, handleOpenTrunk).Wait()
}

func startSendingRandomlyGeneratedTelemetryData() {
	if !telemetrySenderStarted {
		go telemetrySender(stopTelemetrySender, controlTelemetrySender)
		telemetrySenderStarted = true
	}
}

func handleResetDTC(sourceClient mqtt.Client, message mqtt.Message) {
	envelope := protocol.Envelope{}
	if err := json.Unmarshal(message.Payload(), &envelope); err != nil {
		slog.Error("Error during payload unmarshal", "error", err)
		return
	}
	if envelope.Path != "/features/DTC/inbox/messages/resetDTC" {
		slog.Info("wrong path received", "path", envelope.Path)
		return
	}
	slog.Info("CTD Reset DTC received:", "envelope", envelope)

	controlTelemetrySender <- "resetDTC"
}

func handleOpenTrunk(sourceClient mqtt.Client, message mqtt.Message) {
	envelope := protocol.Envelope{}
	if err := json.Unmarshal(message.Payload(), &envelope); err != nil {
		slog.Error("Error during payload unmarshal", "error", err)
		return
	}
	if envelope.Path != "/features/Trunk/inbox/messages/openTrunk" {
		slog.Info("wrong path received", "path", envelope.Path)
		return
	}
	slog.Info("CTD Open Trunk received", "data", currentDeviceData)
	controlTelemetrySender <- "openTrunk"
}

func publishCloudUpdate(feature dittoModel.Feature, featureName string, thingIDSuffix string) {
	fullThingID := getFullThingID(thingIDSuffix)
	command := things.NewCommand(dittoModel.NewNamespacedIDFrom(fullThingID)).
		Twin().
		Feature(featureName).
		Modify(feature)

	envelope := command.Envelope(protocol.WithResponseRequired(false), protocol.WithContentType("application/json"))
	payload, err := json.Marshal(envelope)
	if err != nil {
		slog.Error("Could not marshal the payload", "error", err)
		return
	}

	mqttClient.Publish(fmt.Sprintf("e/%s/%s", currentDeviceData.TenantID, fullThingID), 1, false, payload).Wait()
}

func unsubscribeAllCTD() {
	for _, topic := range ctdSubscriptionTopics {
		slog.Info("Unsubscribing...", "topic", topic)
		mqttClient.Unsubscribe(topic)
	}
	ctdSubscriptionTopics = []string{}
}

func getFullThingID(thingIDSuffix string) string {
	return currentDeviceData.DeviceID + ":" + thingIDSuffix
}

func validateFlags() {
	slog.Info(fmt.Sprintf("Trunk/Alarms State Time: %d seconds", *ctdEffectDelay))
	slog.Info(fmt.Sprintf("Periodic Updates Delay: %d seconds", *periodicUpdatesDelay))
	slog.Info(fmt.Sprintf("Idle State Detection after: %d seconds", *idleStateTime))
	initialPeriodicUpdatesDelay = *periodicUpdatesDelay

	if *periodicUpdatesDelay < 5 {
		panic(fmt.Errorf("periodic-update-delay must be greater or equal to 5 seconds: %d", *periodicUpdatesDelay))
	}
}
