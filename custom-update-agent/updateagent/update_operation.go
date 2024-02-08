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
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/eclipse-kanto/example-applications/custom-update-agent/util"
	"github.com/eclipse-kanto/update-manager/api/types"

	"github.com/rickar/props"
)

type fileAction struct {
	desired *util.File
	current *util.File

	feedbackAction *types.Action
	actionType     util.ActionType
}

type action struct {
	status  types.StatusType
	actions []*fileAction
}

type operation struct {
	temporaryDirectory string
	downloadDirectory  string
	backupDirectory    string

	updateManager *fileUpdateManager
	activityID    string
	desiredState  *internalDesiredState

	allActions *action
}

// UpdateOperation defines an interface for an update operation process
type UpdateOperation interface {
	GetActivityID() string
	Identify() (bool, error)
	Execute(command types.CommandType, baseline string)
	Feedback(status types.StatusType, message string, baseline string)
}

type createUpdateOperation func(*fileUpdateManager, string, *internalDesiredState) UpdateOperation

func newOperation(updMgr *fileUpdateManager, activityID string, desiredState *internalDesiredState) UpdateOperation {
	return &operation{
		updateManager: updMgr,
		activityID:    activityID,
		desiredState:  desiredState,
	}
}

// GetActivityID returns the activity ID associated with this operation
func (o *operation) GetActivityID() string {
	return o.activityID
}

// Identify executes the IDENTIFYING phase, triggered with the full desired state for the domain
func (o *operation) Identify() (bool, error) {
	var err error
	o.temporaryDirectory, err = os.MkdirTemp("", "file_agent")
	if err != nil {
		slog.Error("got error creating temporary directory", "error", err)
		return false, err
	}
	o.downloadDirectory, err = os.MkdirTemp(o.temporaryDirectory, "file_agent_download")
	if err != nil {
		slog.Error("got error creating download directory", "error", err)
		return false, err
	}
	o.backupDirectory, err = os.MkdirTemp(o.temporaryDirectory, "file_agent_backup")
	if err != nil {
		slog.Error("got error creating backup directory", "error", err)
		return false, err
	}

	files, err := os.ReadDir(FileDirectory)
	for _, f := range files {
		filename := f.Name()
		o.copyFile(filename, FileDirectory, o.backupDirectory)
	}
	if err != nil {
		slog.Error("got error reading files directory", "error", err)
		return false, err
	}

	currentFiles := []*util.File{}
	propsFile, err := os.Open(FileDirectory + "/state.props")
	if err != nil {
		slog.Error("got error opening state.props file", "error", err)
		return false, err
	}

	properties, err := props.Read(propsFile)
	if err != nil {
		slog.Error("got error reading state.props file", "error", err)
		return false, err
	}

	for _, filename := range properties.Names() {
		url, _ := properties.Get(filename)
		currentFiles = append(currentFiles, &util.File{Name: filename, DownloadURL: url})
	}

	currentFilesMap := util.AsNamedMap(currentFiles)
	allActions := []*fileAction{}

	slog.Debug("checking desired vs current files")

	for _, desired := range o.desiredState.files {
		filename := desired.Name
		current := currentFilesMap[filename]
		if current != nil {
			delete(currentFilesMap, filename)
		}
		allActions = append(allActions, o.newFileAction(current, desired))
	}

	destroyActions := o.newRemoveActions(currentFilesMap)
	allActions = append(allActions, destroyActions...)

	o.allActions = &action{
		status:  types.StatusIdentified,
		actions: allActions,
	}
	propsFile.Close()

	return len(allActions) > 0, nil
}

func (o *operation) newFileAction(current *util.File, desired *util.File) *fileAction {
	actionType := util.DetermineUpdateAction(current, desired)
	message := util.GetActionMessage(actionType)

	slog.Debug(fmt.Sprintf("[%s] %s", desired.Name, message))

	return &fileAction{
		desired: desired,
		current: current,
		feedbackAction: &types.Action{
			Component: &types.Component{
				ID:      o.updateManager.domainName + ":" + desired.Name,
				Version: o.desiredState.findComponent(desired.Name).Version,
			},
			Status:  types.ActionStatusIdentified,
			Message: message,
		},
		actionType: actionType,
	}
}

func (o *operation) newRemoveActions(toBeRemoved map[string]*util.File) []*fileAction {
	removeActions := []*fileAction{}
	message := util.GetActionMessage(util.ActionRemove)
	for _, current := range toBeRemoved {
		slog.Debug(fmt.Sprintf("[%s] %s", current.Name, message))
		removeActions = append(removeActions, &fileAction{
			desired: nil,
			current: current,
			feedbackAction: &types.Action{
				Component: &types.Component{
					ID: o.updateManager.domainName + ":" + current.Name,
				},
				Status:  types.ActionStatusIdentified,
				Message: message,
			},
			actionType: util.ActionRemove,
		})
	}
	return removeActions
}

// Execute executes each COMMAND (download, update, activate, etc) phase, triggered per baseline or for all the identified actions
func (o *operation) Execute(command types.CommandType, baseline string) {
	commandHandler, action := o.getCommandHandler(baseline, command)
	if action == nil {
		return
	}
	commandHandler(o, action)
}

type commandHandler func(*operation, *action)

var commandHandlers = map[types.CommandType]struct {
	expectedBaselineStatus []types.StatusType
	baselineFailureStatus  types.StatusType
	commandHandler         commandHandler
}{
	types.CommandDownload: {
		expectedBaselineStatus: []types.StatusType{types.StatusIdentified},
		baselineFailureStatus:  types.BaselineStatusDownloadFailure,
		commandHandler:         download,
	},
	types.CommandUpdate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusDownloadSuccess},
		baselineFailureStatus:  types.BaselineStatusUpdateFailure,
		commandHandler:         update,
	},
	types.CommandActivate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusUpdateSuccess},
		baselineFailureStatus:  types.BaselineStatusActivationFailure,
		commandHandler:         activate,
	},
	types.CommandCleanup: {
		baselineFailureStatus: types.BaselineStatusCleanupFailure,
		commandHandler:        cleanup,
	},
}

func (o *operation) getCommandHandler(baseline string, command types.CommandType) (commandHandler, *action) {
	handler, ok := commandHandlers[command]

	if !ok {
		slog.Warn("Ignoring unknown", "command", command)
		return nil, nil
	}
	if o.allActions == nil {
		o.Feedback(handler.baselineFailureStatus, "Unknown baseline "+baseline, baseline)
		return nil, nil
	}
	if len(handler.expectedBaselineStatus) > 0 && !hasStatus(handler.expectedBaselineStatus, o.allActions.status) {
		o.Feedback(handler.baselineFailureStatus, fmt.Sprintf("%s is possible only after status %s is reported", command, asStatusString(handler.expectedBaselineStatus)), baseline)
		return nil, nil
	}
	return handler.commandHandler, o.allActions
}

// ActionAdd and ActionReplace: download file from defined url to temporary file directory.
func download(o *operation, baselineAction *action) {
	var lastAction *fileAction
	var lastActionErr error
	lastActionMessage := ""

	slog.Debug("downloading - starting...")
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadSuccess, lastAction, types.ActionStatusDownloadSuccess, lastActionMessage)

		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadFailure, lastAction, types.ActionStatusDownloadFailure, lastActionErr.Error())
			rollback(o, baselineAction)
		}

		slog.Debug("downloading - done.")
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if lastAction != nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloading, lastAction, types.ActionStatusDownloadSuccess, lastActionMessage)
		}
		lastAction = action

		if action.actionType == util.ActionAdd || action.actionType == util.ActionReplace {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloading, action, types.ActionStatusDownloading, action.feedbackAction.Message)

			if err := o.downloadFile(action.desired); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "New file added."
		} else {
			lastAction = nil
		}
	}
}

// ActionAdd, ActionNone and ActionReplace: update the state.props file with the new file-dowload url pairs.
func activate(o *operation, baselineAction *action) {
	var lastAction *fileAction
	var lastActionErr error

	lastActionMessage := ""

	propsFile, err := os.OpenFile(FileDirectory+"/state.props", os.O_RDWR, 0666)

	if err != nil {
		slog.Error("got error opening state.props file", "error", err)
		return
	}

	propsFile.Truncate(0)
	defer propsFile.Close()

	slog.Debug("activating - starting...")

	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationSuccess, lastAction, types.ActionStatusActivationSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationFailure, lastAction, types.ActionStatusActivationFailure, lastActionErr.Error())
			rollback(o, baselineAction)
		}

		slog.Debug("activating - done.")
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if lastAction != nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivating, lastAction, types.ActionStatusActivationSuccess, lastActionMessage)
		}

		lastAction = action
		if action.actionType == util.ActionAdd || action.actionType == util.ActionReplace || action.actionType == util.ActionNone {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivating, action, types.ActionStatusActivating, action.feedbackAction.Message)
			lastActionMessage = "Desired file added to state.props file."
			if err := addProperty(action.desired.Name, action.desired.DownloadURL); err != nil {
				lastActionErr = err
				slog.Error("got error updating state.props file", "error", err)
				return
			}
		} else {
			lastAction = nil
		}
	}
}

// ActionAdd, ActionReplace: move file from temporary directory to fileagent directory.
func update(o *operation, baselineAction *action) {
	var lastAction *fileAction
	var lastActionErr error
	lastActionMessage := ""

	slog.Debug("updating - starting...")
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateSuccess, lastAction, types.ActionStatusUpdateSuccess, lastActionMessage)
		} else {
			slog.Debug("last action error")
			slog.Debug(lastActionErr.Error())
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateFailure, lastAction, types.ActionStatusUpdateFailure, lastActionErr.Error())
			rollback(o, baselineAction)
		}

		slog.Debug("updating - done.")
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if lastAction != nil {
			lastAction.feedbackAction.Status = types.ActionStatusUpdateSuccess
			lastAction.feedbackAction.Message = lastActionMessage
		}
		lastAction = action
		if action.actionType == util.ActionAdd || action.actionType == util.ActionReplace {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdating, action, types.ActionStatusUpdating, action.feedbackAction.Message)
			if err := o.copyFile(action.desired.Name, o.downloadDirectory, FileDirectory); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "File added to directory."
		} else if action.actionType == util.ActionRemove {
			if err := o.removeFile(action.current); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "File removed from directory."
		} else {
			lastActionMessage = action.feedbackAction.Message
		}
	}
}

// Restores fileagent directory backup and removes all files in temporary directory
func rollback(o *operation, baselineAction *action) {
	var lastAction *fileAction
	var lastActionMessage string
	var lastActionErr error

	slog.Debug("rollback - starting...")

	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackSuccess, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackFailure, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		}
		slog.Debug("rollback - done.")
	}()

	files, err := os.ReadDir(FileDirectory)
	if err != nil {
		slog.Error("got error reading directory", "error", err)
		lastActionErr = err
		return
	}
	backupFiles, err := os.ReadDir(o.backupDirectory)
	if err != nil {
		slog.Error("got error opening directory", "error", err)
		lastActionErr = err
		return
	}
	if !reflect.DeepEqual(files, backupFiles) {
		for _, entry := range files {
			err = os.Remove(FileDirectory + "/" + entry.Name())
			if err != nil {
				slog.Error("got error removing file", "error", err)
				lastActionErr = err
				return
			}
		}

		for _, f := range files {
			filename := f.Name()
			o.copyFile(filename, o.backupDirectory, FileDirectory)
		}
	}
}

// ActionRemove: removes the old file from fileagent directory.
// ActionAdd and ActionReplace: removes temporary download directory.
func cleanup(o *operation, baselineAction *action) {
	slog.Debug("cleanup - starting...")

	o.cleanupTemporaryFolders()
	o.Feedback(types.BaselineStatusCleanupSuccess, "", "")

	slog.Debug("cleanup - done.")
}

// Feedback sends desired state feedback responses, baseline parameter is optional
func (o *operation) Feedback(status types.StatusType, message string, baseline string) {
	o.updateManager.eventCallback.HandleDesiredStateFeedbackEvent(o.updateManager.domainName, o.activityID, baseline, status, message, o.toFeedbackActions())
}

func (o *operation) updateBaselineActionStatus(baseline *action, baselineStatus types.StatusType,
	action *fileAction, actionStatus types.ActionStatusType, message string) {
	if action != nil {
		action.feedbackAction.Status = actionStatus
		action.feedbackAction.Message = message
	}
	baseline.status = baselineStatus
	o.Feedback(baselineStatus, "", "")
}

func (o *operation) toFeedbackActions() []*types.Action {
	if o.allActions == nil {
		return nil
	}
	result := make([]*types.Action, len(o.allActions.actions))
	for i, action := range o.allActions.actions {
		result[i] = action.feedbackAction
	}
	return result
}

func hasStatus(where []types.StatusType, what types.StatusType) bool {
	for _, status := range where {
		if status == what {
			return true
		}
	}
	return false
}

func (o *operation) copyFile(filename string, sourcePath string, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath + "/" + filename)
	if err != nil {
		slog.Error(fmt.Sprintf("got error opening file [%s]", filename), "error", err)
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(destinationPath + "/" + filename)
	if err != nil {
		slog.Error(fmt.Sprintf("got error creating file [%s]", filename), "error", err)
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		slog.Error(fmt.Sprintf("got error copying file [%s] to [%s]", filename, FileDirectory), "error", err)
		return err

	}
	return err
}

func (o *operation) cleanupTemporaryFolders() error {
	err := os.RemoveAll(o.temporaryDirectory)
	if err != nil {
		slog.Error("got error removing temporary folders", "error", err)
	}
	return err
}

func (o *operation) removeFile(desired *util.File) error {
	err := os.Remove(FileDirectory + "/" + desired.Name)
	if err != nil {
		slog.Error(fmt.Sprintf("got error removing file [%s]", desired.Name), "error", err)
	}
	return err
}

func (o *operation) downloadFile(desired *util.File) error {
	resp, err := http.Get(desired.DownloadURL)
	if err != nil {
		slog.Debug(fmt.Sprintf("could not download file from url [%s]", desired.DownloadURL), "error", err)
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(o.downloadDirectory + "/" + desired.Name)

	if err != nil {
		slog.Debug(fmt.Sprintf("could not create file [%s]", desired.Name), "error", err)
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("could not copy contents from [%s] to file [%s]", desired.DownloadURL, desired.Name), "error", err)
		return err
	}

	return nil
}

func addProperty(key string, value string) error {
	propsFilePath := FileDirectory + "/state.props"
	propsFile, err := os.OpenFile(propsFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		slog.Debug(fmt.Sprintf("could not open file [%s]", propsFilePath), "error", err)
		return err
	}
	defer propsFile.Close()

	w := bufio.NewWriter(propsFile)

	newProperties := props.NewProperties()
	newProperties.Set(key, value)
	newProperties.Write(w)

	err = w.Flush()
	return err
}

func asStatusString(what []types.StatusType) string {
	var sb strings.Builder
	for _, status := range what {
		if sb.Len() > 0 {
			sb.WriteRune('|')
		}
		sb.WriteString(string(status))
	}
	return sb.String()
}
