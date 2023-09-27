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
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	dittoModel "github.com/eclipse/ditto-clients-golang/model"
)

var OBDFeature = dittoModel.Feature{}
var trunkFeature = dittoModel.Feature{}

type TelemetryData struct {
	ControlModuleVoltage string
	B1601                bool
	B1602                bool
	B1603                bool
	isTrunkOpen          bool
}

var telemetry TelemetryData

var dtcStateStartTime time.Time   // device trouble codes start time to store the state
var trunkStateStartTime time.Time // trunk start time to store the state
var systemStartTime time.Time     // used by the idle state machine to detect when the simulator enters idle state. Gets reset upon receiving CTD command
var trunkOpenState = true         // holds the difference in trunk state. Will send telemetry only on state changed

func telemetrySender(stopTelemetrySender <-chan struct{}, controlChan <-chan string) {
	resetIdleState()
	telemetry.isTrunkOpen = false
	processTrunkFeatureData() // send initial trunk state - closed
	processOBDfeatureData()
	for {
		detectIdleState()
		select {
		case <-time.After(time.Duration(*periodicUpdatesDelay) * time.Second): // Introduce a periodic updates delay
			processTrunkFeatureData()
			processOBDfeatureData()
		case <-stopTelemetrySender: // If a signal is received on the stopChan, stop the telemetry sender
			slog.Info("Stopping Telemetry Sender...")
			return
		case signal := <-controlChan:
			switch signal {
			case "resetDtc":
				slog.Info("Resetting DTC...")
				resetIdleState()
				dtcStateStartTime = time.Now()
				telemetry.B1601 = false
				telemetry.B1602 = false
				telemetry.B1603 = false
				processOBDfeatureData()
			case "openTrunk":
				slog.Info("Opening the Trunk...")
				resetIdleState()
				trunkStateStartTime = time.Now()
				telemetry.isTrunkOpen = true
				processTrunkFeatureData()
			default:
				slog.Info("Unknown control signal: ", signal)
			}
		}
	}
}

func processTrunkFeatureData() {
	elapsedTrunkStateTime := time.Since(trunkStateStartTime)
	if !(elapsedTrunkStateTime > time.Duration(*ctdEffectDelay)*time.Second) {
		telemetry.isTrunkOpen = true // keep the trunk state for a configured time window
	} else {
		telemetry.isTrunkOpen = false // reset the state always to "closed" after state expired
	}

	if trunkOpenState != telemetry.isTrunkOpen { // send telemetry data only when state changed
		sendTrunkTelemetry()
		trunkOpenState = telemetry.isTrunkOpen
	}
}

func processOBDfeatureData() {
	elapsedDtcStateTime := time.Since(dtcStateStartTime)

	if elapsedDtcStateTime > time.Duration(*ctdEffectDelay)*time.Second {
		telemetry.B1601 = rand.Intn(10) > 5
		telemetry.B1602 = rand.Intn(10) > 5
		telemetry.B1603 = rand.Intn(10) > 5
	}

	voltage := 9.0 + rand.Float64()*6.0
	formattedVoltage := fmt.Sprintf("%.2f", voltage)
	telemetry.ControlModuleVoltage = formattedVoltage

	sendOBDTelemetry()
}

func sendTrunkTelemetry() {
	trunkFeature.WithProperties(map[string]interface{}{
		"isOpen": telemetry.isTrunkOpen,
	})
	slog.Info("Sending Telemetry Data", "trunkFeature", trunkFeature)
	publishCloudUpdate(trunkFeature, "trunk", "Vehicle")
}

func sendOBDTelemetry() {
	OBDFeature.WithProperties(map[string]interface{}{
		"controlModuleVoltage": telemetry.ControlModuleVoltage,
		"b1601":                telemetry.B1601,
		"b1602":                telemetry.B1602,
		"b1603":                telemetry.B1603,
	})

	slog.Info("Sending Telemetry Data", "OBDFeature", OBDFeature)
	publishCloudUpdate(OBDFeature, "OBD", "Vehicle")
}

func resetIdleState() {
	slog.Info("Resetting idle state...")
	*periodicUpdatesDelay = initialPeriodicUpdatesDelay
	systemStartTime = time.Now()
}

func detectIdleState() {
	elapsedStartTime := time.Since(systemStartTime)
	slog.Info(fmt.Sprintf("Elapsed time in seconds since active: %.2f seconds\n", elapsedStartTime.Seconds()))
	if elapsedStartTime > time.Duration(*idleStateTime)*time.Second {
		slog.Info("Detected idle state, activating exponential back-off")
		*periodicUpdatesDelay += *periodicUpdatesDelay
		slog.Info(fmt.Sprintf("Periodic updates delay set to: %d seconds \n", *periodicUpdatesDelay))
	}
}
