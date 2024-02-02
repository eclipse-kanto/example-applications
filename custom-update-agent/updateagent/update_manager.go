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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/version"
	"github.com/eclipse-kanto/example-applications/custom-update-agent/util"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/types"

	"github.com/rickar/props"
)

const (
	updateManagerName = "Eclipse Kanto File Update Agent"
	parameterDomain   = "domain"
)

// FileDirectory points to the directory managed by the Files Update Agent
var FileDirectory = ""

type fileUpdateManager struct {
	domainName string

	applyLock             sync.Mutex
	eventCallback         api.UpdateManagerCallback
	createUpdateOperation createUpdateOperation
	operation             UpdateOperation
}

// Name returns the name of this update manager, e.g. "files".
func (updMgr *fileUpdateManager) Name() string {

	return updMgr.domainName
}

// Apply triggers the update operation with the given activity ID and desired state with files.
// First, it validates the received desired state specification and identifies the actions to be applied.
// If errors are detected, then IDENTIFICATION_FAILED feedback status is reported and operation finishes unsuccessfully.
// Otherwise, IDENTIFIED feedback status with identified actions is reported and it will wait for further commands to proceed.
func (updMgr *fileUpdateManager) Apply(ctx context.Context, activityID string, desiredState *types.DesiredState) {
	updMgr.applyLock.Lock()
	defer updMgr.applyLock.Unlock()

	log.Debug("processing desired state - start")
	// create operation instance
	internalDesiredState, err := toInternalDesiredState(desiredState, updMgr.domainName)
	if err != nil {
		log.ErrorErr(err, "could not parse desired state components as file configurations")
		updMgr.eventCallback.HandleDesiredStateFeedbackEvent(updMgr.Name(), activityID, "", types.StatusIdentificationFailed, err.Error(), []*types.Action{})
		return
	}
	newOperation := updMgr.createUpdateOperation(updMgr, activityID, internalDesiredState)

	// identification phase
	newOperation.Feedback(types.StatusIdentifying, "", "")
	hasActions, err := newOperation.Identify()
	if err != nil {
		newOperation.Feedback(types.StatusIdentificationFailed, err.Error(), "")
		log.ErrorErr(err, "processing desired state - identification phase failed")
		return
	}
	newOperation.Feedback(types.StatusIdentified, "", "")
	if !hasActions {
		log.Debug("processing desired state - identification phase completed, no actions identified, sending COMPLETE status")
		newOperation.Feedback(types.StatusCompleted, "", "")
		return
	}
	updMgr.operation = newOperation
	log.Debug("processing desired state - identification phase completed, waiting for commands...")
}

// Command processes received desired state command.
func (updMgr *fileUpdateManager) Command(ctx context.Context, activityID string, command *types.DesiredStateCommand) {
	if command == nil {
		log.Error("Skipping received command for activityId %s, but no payload.", activityID)
		return
	}
	updMgr.applyLock.Lock()
	defer updMgr.applyLock.Unlock()

	operation := updMgr.operation
	if operation == nil {
		log.Warn("Ignoring received command %s for baseline %s and activityId %s, but no operation in progress.", command.Command, command.Baseline, activityID)
		return
	}
	if operation.GetActivityID() != activityID {
		log.Warn("Ignoring received command %s for baseline %s and activityId %s, but not matching operation in progress [%s].",
			command.Command, command.Baseline, activityID, operation.GetActivityID())
		return
	}
	operation.Execute(command.Command, command.Baseline)
}

// Get returns the current state as an inventory graph.
// The inventory graph includes a root software node (type APPLICATION) representing the update agent itself and a list of software nodes (type DATA) representing the available files.
func (updMgr *fileUpdateManager) Get(ctx context.Context, activityID string) (*types.Inventory, error) {
	return toInventory(updMgr.asSoftwareNode(), updMgr.getCurrentFiles()), nil
}

func toInventory(swNodeAgent *types.SoftwareNode, swNodeFiles []*types.SoftwareNode) *types.Inventory {
	swNodes := []*types.SoftwareNode{swNodeAgent}
	associations := []*types.Association{}
	if len(swNodeFiles) > 0 {
		swNodes = append(swNodes, swNodeFiles...)
		for _, swNodeContainer := range swNodeFiles {
			swNodeContainer.ID = swNodeAgent.Parameters[0].Value + ":" + swNodeContainer.ID

			associations = append(associations, &types.Association{
				SourceID: swNodeAgent.ID,
				TargetID: swNodeContainer.ID,
			})
		}
	}
	return &types.Inventory{
		SoftwareNodes: swNodes,
		Associations:  associations,
	}
}

func (updMgr *fileUpdateManager) asSoftwareNode() *types.SoftwareNode {
	return &types.SoftwareNode{
		InventoryNode: types.InventoryNode{
			ID:      updMgr.Name() + "-update-agent",
			Version: version.ProjectVersion,
			Name:    updateManagerName,
			Parameters: []*types.KeyValuePair{
				{
					Key:   parameterDomain,
					Value: updMgr.Name(),
				},
			},
		},
		Type: types.SoftwareTypeApplication,
	}
}

func (updMgr *fileUpdateManager) getCurrentFiles() []*types.SoftwareNode {
	files := []*util.File{}
	propsFilePath := FileDirectory + "/state.props"

	_, err := os.Stat(propsFilePath)

	if errors.Is(err, os.ErrNotExist) {
		entries, err := os.ReadDir(FileDirectory)
		if err != nil {
			slog.Error("got error checking current files", "error", err)
			return nil

		}
		_, err = os.Create(propsFilePath)
		if err != nil {
			slog.Error(fmt.Sprintf("got error creating file [%s]", "state.props"), "error", err)

		}
		for _, entry := range entries {
			addProperty(entry.Name(), "unknown")
		}
	}
	propsFile, err := os.Open(propsFilePath)
	if err != nil {
		slog.Error("got error checking current files", "error", err)
		return nil

	}

	properties, err := props.Read(propsFile)
	if err != nil {

		slog.Error("got error when reading state.props file", "error", err)
		return nil

	}

	for _, filename := range properties.Names() {
		url, _ := properties.Get(filename)
		files = append(files, &util.File{Name: filename, DownloadURL: url})
	}

	return util.FromFiles(files)
}

// Dispose releases all resources used by this instance
func (updMgr *fileUpdateManager) Dispose() error {
	return nil
}

// WatchEvents subscribes for events that update the current state inventory
func (updMgr *fileUpdateManager) WatchEvents(ctx context.Context) {
	// no container events handled yet - current state inventory reported only on initial start or explicit get request
}

// SetCallback sets the callback instance that is used for desired state feedback / current state notifications.
// It is set when the update agent instance is started
func (updMgr *fileUpdateManager) SetCallback(callback api.UpdateManagerCallback) {
	updMgr.eventCallback = callback
}
