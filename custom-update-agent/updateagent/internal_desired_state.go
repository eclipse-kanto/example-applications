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
	"fmt"

	"github.com/eclipse-kanto/example-applications/custom-update-agent/util"

	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/pkg/errors"
)

type internalDesiredState struct {
	desiredState *types.DesiredState
	files        []*util.File
}

func (ds *internalDesiredState) findComponent(name string) types.Component {
	for _, component := range ds.desiredState.Domains[0].Components {
		if component.ID == name {
			return component.Component
		}
	}
	return types.Component{}
}

// toInternalDesiredState converts incoming desired state into an internal desired state structure
func toInternalDesiredState(desiredState *types.DesiredState, domainName string) (*internalDesiredState, error) {
	if len(desiredState.Domains) != 1 {
		return nil, fmt.Errorf("one domain expected in desired state specification, but got %d", len(desiredState.Domains))
	}
	if desiredState.Domains[0].ID != domainName {
		return nil, fmt.Errorf("domain id mismatch - expecting %s, received %s", domainName, desiredState.Domains[0].ID)
	}
	files, err := util.ToFiles(desiredState.Domains[0].Components)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert desired state components to container configurations")
	}

	return &internalDesiredState{
		desiredState: desiredState,
		files:        files,
	}, nil
}
