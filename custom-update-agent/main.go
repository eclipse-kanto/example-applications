// Copyright (c) 2024 Contributors to the Eclipse Foundation
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
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse-kanto/example-applications/custom-update-agent/updateagent"
	"github.com/eclipse-kanto/example-applications/custom-update-agent/util"

	"github.com/eclipse-kanto/update-manager/api"
)

const (

	// default local connection config
	connectionBrokerURLDefault         = "tcp://localhost:1883"
	connectionKeepAliveDefault         = "20s"
	connectionDisconnectTimeoutDefault = "250ms"
	connectionClientUsername           = ""
	connectionClientPassword           = ""
	connectTimeoutTimeoutDefault       = "30s"
	acknowledgeTimeoutDefault          = "15s"
	subscribeTimeoutDefault            = "15s"
	unsubscribeTimeoutDefault          = "5s"

	// default update agent config
	updateAgentEnableDefault                 = false
	updateAgentDomainDefault                 = "files"
	updateAgentVerboseInventoryReportDefault = false
)

func parseDuration(duration string) time.Duration {
	d, _ := time.ParseDuration(duration)
	return d
}

func getDefaultCustomUpdateAgentOpts() []updateagent.FileUpdateAgentOpt {
	updateAgentOpts := []updateagent.FileUpdateAgentOpt{}

	updateAgentOpts = append(updateAgentOpts,
		updateagent.WithDomainName(updateAgentDomainDefault),
		updateagent.WithConnectionBroker(connectionBrokerURLDefault),
		updateagent.WithConnectionKeepAlive(parseDuration(connectionKeepAliveDefault)),
		updateagent.WithConnectionDisconnectTimeout(parseDuration(connectionDisconnectTimeoutDefault)),
		updateagent.WithConnectionClientUsername(connectionClientUsername),
		updateagent.WithConnectionClientPassword(connectionClientPassword),
		updateagent.WithConnectionConnectTimeout(parseDuration(connectTimeoutTimeoutDefault)),
		updateagent.WithConnectionAcknowledgeTimeout(parseDuration(acknowledgeTimeoutDefault)),
		updateagent.WithConnectionSubscribeTimeout(parseDuration(subscribeTimeoutDefault)),
		updateagent.WithConnectionUnsubscribeTimeout(parseDuration(unsubscribeTimeoutDefault)))

	return updateAgentOpts
}

func main() {

	//set logger level and output
	logger := util.ConfigLogger(slog.LevelDebug, os.Stdout)
	slog.SetDefault(&logger)

	updateAgent, _ := updateagent.Init(getDefaultCustomUpdateAgentOpts())

	err := updateAgent.(api.UpdateAgent).Start(context.Background())
	if err != nil {
		slog.Error("could not start Update Agent service! got", "error", err)
	} else {
		slog.Info("successfully started Update Agent service")
	}

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-signalChan
	slog.Info("Exiting!, recieved", "signal", sig)
	updateAgent.(api.UpdateAgent).Stop()

}
