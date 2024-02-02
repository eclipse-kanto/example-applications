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
	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/agent"
	"github.com/eclipse-kanto/update-manager/mqtt"
)

func newUpdateAgent(
	domainName string,
	broker string,
	keepAlive string,
	disconnectTimeout string,
	clientUsername string,
	clientPassword string,
	connectTimeout string,
	acknowledgeTimeout string,
	subscribeTimeout string,
	unsubscribeTimeout string,
) (api.UpdateAgent, error) {

	mqttClient, _ := mqtt.NewUpdateAgentClient(domainName, &mqtt.ConnectionConfig{
		Broker:             broker,
		KeepAlive:          keepAlive,
		DisconnectTimeout:  disconnectTimeout,
		Username:           clientUsername,
		Password:           clientPassword,
		ConnectTimeout:     connectTimeout,
		AcknowledgeTimeout: acknowledgeTimeout,
		SubscribeTimeout:   subscribeTimeout,
		UnsubscribeTimeout: unsubscribeTimeout,
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

// Init initializes a new Update Agent instance using given configuration and domain
func Init(config mqtt.ConnectionConfig, domain string) (interface{}, error) {

	return newUpdateAgent(
		domain,
		config.Broker,
		config.KeepAlive,
		config.DisconnectTimeout,
		config.Username,
		config.Password,
		config.ConnectTimeout,
		config.AcknowledgeTimeout,
		config.SubscribeTimeout,
		config.UnsubscribeTimeout,
	)
}
