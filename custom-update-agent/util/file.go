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

// File represents the file instance in directory
type File struct {
	Name        string `json:"file_name"`
	DownloadURL string `json:"download_url"`
}

// AsNamedMap returns a map of file where key is the file's name
func AsNamedMap(filesList []*File) map[string]*File {
	result := map[string]*File{}
	for _, file := range filesList {
		if file != nil {
			result[file.Name] = file
		}
	}
	return result
}
