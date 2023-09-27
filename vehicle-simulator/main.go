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
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

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

type DeviceData struct {
	DeviceID string `json:"deviceID"`
	TenantID string `json:"tenantID"`
	PolicyID string `json:"policyID"`
}

var mqttClient mqtt.Client
var deviceData = DeviceData{}
var telemetrySenderStarted bool
var selectedConnectorDeviceID string // holds the last deviceID from the previous active connector
var ctdSubscriptionTopics []string

var stopTelemetrySender = make(chan struct{})
var controlTelemetrySender = make(chan string)

// Define flag variables
var ctdEffectDelay = flag.Int("ctd-effect-delay", 30, "In seconds, Configure trunk/alarms state time")
var periodicUpdatesDelay = flag.Int("periodic-updates-delay", 5, "In seconds, Configure periodic updates (minimum 5)")
var idleStateTime = flag.Int("idle-state-time", 120, "In seconds, Configure when the system will enter idle state. Idle state will send telemetry data infrequently. Receiving CTD command will reset the idle state.")
var initialPeriodicUpdatesDelay = 0 // set the initial periodic updates delay in case of idle state active

func main() {
	flag.Parse()
	validateFlags()

	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	initializeClient()
	getDeviceData()

	defer func() { stopTelemetrySender <- struct{}{} }()
	defer mqttClient.Disconnect(1000)

	<-keepAlive
}

func initializeClient() {
	brokerURL := fmt.Sprintf("tcp://%s:%s", getEnvVariable("ADAPTER_BROKER_ADDRESS", "localhost"), getEnvVariable("ADAPTER_BROKER_PORT", "1883"))
	slog.Info("Broker configuration", "URL", brokerURL)
	options := mqtt.NewClientOptions()
	options.AddBroker(brokerURL)
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
	if err := json.Unmarshal(message.Payload(), &deviceData); err != nil {
		slog.Error("Error during payload unmarshal: ", err)
		return
	}

	slog.Info("Connector Device Data received successfully", "data", deviceData)
	if deviceData.DeviceID == selectedConnectorDeviceID {
		slog.Info("Skipping the same connector deviceID instance.")
		return
	}

	unsubscribeAllCTD()
	subscribeForCTD(deviceData.DeviceID)
	startSendingRandomlyGeneratedTelemetryData()
	selectedConnectorDeviceID = deviceData.DeviceID
}

func subscribeForCTD(deviceId string) {
	slog.Info("Subscribing for CTD commands: resetDtc & openTrunk...")
	topicResetDtc := fmt.Sprintf("command//%s:Vehicle/req//resetDtc", deviceId)
	ctdSubscriptionTopics = append(ctdSubscriptionTopics, topicResetDtc)
	topicOpenTrunk := fmt.Sprintf("command//%s:Vehicle/req//openTrunk", deviceId)
	ctdSubscriptionTopics = append(ctdSubscriptionTopics, topicOpenTrunk)

	mqttClient.Subscribe(topicResetDtc, 1, handleResetDtc).Wait()
	mqttClient.Subscribe(topicOpenTrunk, 1, handleOpenTrunk).Wait()
}

func startSendingRandomlyGeneratedTelemetryData() {
	if !telemetrySenderStarted {
		go telemetrySender(stopTelemetrySender, controlTelemetrySender)
		telemetrySenderStarted = true
	}
}

func handleResetDtc(sourceClient mqtt.Client, message mqtt.Message) {
	if err := json.Unmarshal(message.Payload(), &deviceData); err != nil {
		slog.Error("Error during payload unmarshal", "error", err)
		return
	}

	slog.Info("CTD Reset DTC received:", deviceData)

	controlTelemetrySender <- "resetDtc"
}

func handleOpenTrunk(sourceClient mqtt.Client, message mqtt.Message) {
	if err := json.Unmarshal(message.Payload(), &deviceData); err != nil {
		slog.Error("Error during payload unmarshal", "error", err)
		return
	}
	slog.Info("CTD Open Trunk received", "data", deviceData)
	controlTelemetrySender <- "openTrunk"
}

func publishCloudUpdate(feature dittoModel.Feature, featureName string, thingIdSuffix string) {
	fullThingID := getFullThingId(thingIdSuffix)
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

	mqttClient.Publish(fmt.Sprintf("e/%s/%s", deviceData.TenantID, fullThingID), 1, false, payload).Wait()
}

func unsubscribeAllCTD() {
	for _, topic := range ctdSubscriptionTopics {
		slog.Info("Unsubscribing...", "topic", topic)
		mqttClient.Unsubscribe(topic)
	}
	ctdSubscriptionTopics = []string{}
}

func getFullThingId(thingIdSuffix string) string {
	return deviceData.DeviceID + ":" + thingIdSuffix
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
