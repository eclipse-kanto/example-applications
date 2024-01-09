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
)

// FileUpdateAgentOpt represents the available configuration options for the Containers UpdateAgent service
type FileUpdateAgentOpt func(updateAgentOptions *updateAgentOpts) error

type updateAgentOpts struct {
	domainName         string
	broker             string
	keepAlive          time.Duration
	disconnectTimeout  time.Duration
	clientUsername     string
	clientPassword     string
	connectTimeout     time.Duration
	acknowledgeTimeout time.Duration
	subscribeTimeout   time.Duration
	unsubscribeTimeout time.Duration
}

func applyOptsUpdateAgent(updateAgentOpts *updateAgentOpts, opts ...FileUpdateAgentOpt) error {
	for _, o := range opts {
		if err := o(updateAgentOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithDomainName configures the domain name for the containers update agent
func WithDomainName(domain string) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.domainName = domain
		return nil
	}
}

// WithConnectionBroker configures the broker, where the connection will be established
func WithConnectionBroker(broker string) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.broker = broker
		return nil
	}
}

// WithConnectionKeepAlive configures the time between between each check for the connection presence
func WithConnectionKeepAlive(keepAlive time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.keepAlive = keepAlive
		return nil
	}
}

// WithConnectionDisconnectTimeout configures the duration of inactivity before disconnecting from the broker
func WithConnectionDisconnectTimeout(disconnectTimeout time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.disconnectTimeout = disconnectTimeout
		return nil
	}
}

// WithConnectionClientUsername configures the client username used when establishing connection to the broker
func WithConnectionClientUsername(username string) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.clientUsername = username
		return nil
	}
}

// WithConnectionClientPassword configures the client password used when establishing connection to the broker
func WithConnectionClientPassword(password string) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.clientPassword = password
		return nil
	}
}

// WithConnectionConnectTimeout configures the timeout before terminating the connect attempt
func WithConnectionConnectTimeout(connectTimeout time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.connectTimeout = connectTimeout
		return nil
	}
}

// WithConnectionAcknowledgeTimeout configures the timeout for the acknowledge receival
func WithConnectionAcknowledgeTimeout(acknowledgeTimeout time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.acknowledgeTimeout = acknowledgeTimeout
		return nil
	}
}

// WithConnectionSubscribeTimeout configures the timeout before terminating the subscribe attempt
func WithConnectionSubscribeTimeout(subscribeTimeout time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.subscribeTimeout = subscribeTimeout
		return nil
	}
}

// WithConnectionUnsubscribeTimeout configures the timeout before terminating the unsubscribe attempt
func WithConnectionUnsubscribeTimeout(unsubscribeTimeout time.Duration) FileUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.unsubscribeTimeout = unsubscribeTimeout
		return nil
	}
}
