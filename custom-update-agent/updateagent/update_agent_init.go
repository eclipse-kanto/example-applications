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

package updateagent

import (
	"time"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/agent"
	"github.com/eclipse-kanto/update-manager/mqtt"
)

func newUpdateAgent(
	domainName string,
	broker string,
	keepAlive time.Duration,
	disconnectTimeout time.Duration,
	clientUsername string,
	clientPassword string,
	connectTimeout time.Duration,
	acknowledgeTimeout time.Duration,
	subscribeTimeout time.Duration,
	unsubscribeTimeout time.Duration,
) (api.UpdateAgent, error) {

	mqttClient := mqtt.NewUpdateAgentClient(domainName, &mqtt.ConnectionConfig{
		Broker:             broker,
		KeepAlive:          keepAlive.Milliseconds(),
		DisconnectTimeout:  disconnectTimeout.Milliseconds(),
		Username:           clientUsername,
		Password:           clientPassword,
		ConnectTimeout:     connectTimeout.Milliseconds(),
		AcknowledgeTimeout: acknowledgeTimeout.Milliseconds(),
		SubscribeTimeout:   subscribeTimeout.Milliseconds(),
		UnsubscribeTimeout: unsubscribeTimeout.Milliseconds(),
	})

	return agent.NewUpdateAgent(mqttClient, newUpdateManager(domainName)), nil
}

// newUpdateManager instantiates a new update manager instance
func newUpdateManager(domainName string) api.UpdateManager {
	return &fileUpdateManager{
		domainName: domainName,

		createUpdateOperation: newOperation,
	}
}
func Init(opts []FileUpdateAgentOpt) (interface{}, error) {
	uaOpts := &updateAgentOpts{}
	if err := applyOptsUpdateAgent(uaOpts, opts...); err != nil {
		return nil, err
	}

	return newUpdateAgent(
		uaOpts.domainName,
		uaOpts.broker,
		uaOpts.keepAlive,
		uaOpts.disconnectTimeout,
		uaOpts.clientUsername,
		uaOpts.clientPassword,
		uaOpts.connectTimeout,
		uaOpts.acknowledgeTimeout,
		uaOpts.subscribeTimeout,
		uaOpts.unsubscribeTimeout,
	)
}
