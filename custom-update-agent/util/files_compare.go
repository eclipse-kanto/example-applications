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

package util

import (
	"fmt"
)

// ActionType defines a type for an action to achieve desired file state
type ActionType int

const (
	// ActionNone denotes that the current file already has the desired configuration, no action is required
	ActionNone ActionType = iota
	// ActionAdd denotes that a new file with desired configurtion shall be downloaded and added to directory
	ActionAdd
	// ActionReplace denotes the existing file shall be replaced by a new file with desired configurtionadded in directory
	ActionReplace
	// ActionRemove denotes that the existing file shall be removed from directory
	ActionRemove
)

// DetermineUpdateAction compares the current file with the desired one and determines what action shall be done to achieve desired state
func DetermineUpdateAction(current *File, desired *File) ActionType {
	if current == nil {
		return ActionAdd
	}
	if current.DownloadURL != desired.DownloadURL && current.Name == desired.Name {
		return ActionReplace
	}
	return ActionNone
}

// GetActionMessage returns a text message describing the given action type
func GetActionMessage(actionType ActionType) string {
	switch actionType {
	case ActionNone:
		return "No changes detected, file will remain in directory with current state."
	case ActionAdd:
		return "New file will be downloaded and added in direcotry."
	case ActionReplace:
		return "Existing file will be replaced by a new one."
	case ActionRemove:
		return "Existing file will be removed, no longer needed."
	}
	return "Unknown action type: " + fmt.Sprint(actionType)
}
