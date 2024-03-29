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
	"github.com/pkg/errors"
)

// ToFiles converts a list of components into a list of files
func ToFiles(components []*types.ComponentWithConfig) ([]*File, error) {
	files := []*File{}
	for _, component := range components {
		file, err := toFile(component)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid configuration for container %s", component.ID)
		}
		files = append(files, file)
	}
	return files, nil
}

func toFile(component *types.ComponentWithConfig) (*File, error) {
	file := &File{}
	for _, kvPair := range component.Config {
		if kvPair.Key == "file_name" {
			file.Name = kvPair.Value
		}
		if kvPair.Key == "download_url" {
			file.DownloadURL = kvPair.Value
		}
	}
	return file, nil
}
