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
	"github.com/eclipse-kanto/update-manager/api/types"
)

// FromFiles turns a list of files into a list of software nodes
func FromFiles(files []*File) []*types.SoftwareNode {
	softwareNodes := make([]*types.SoftwareNode, len(files))
	for i, file := range files {
		softwareNodes[i] = fromFile(file)
	}
	return softwareNodes
}

func fromFile(file *File) *types.SoftwareNode {
	params := []*types.KeyValuePair{}

	params = append(params, &types.KeyValuePair{Key: "download_url", Value: file.DownloadURL})

	return &types.SoftwareNode{
		InventoryNode: types.InventoryNode{
			ID:         file.Name,
			Parameters: params,
		},
		Type: types.SoftwareTypeData,
	}
}
