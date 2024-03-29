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
	"log/slog"
	"os"
)

// ConfigLogger is used for configuring the slog logger log level
func ConfigLogger(level slog.Level, output *os.File) slog.Logger {
	opts := slog.HandlerOptions{
		Level: level,
	}

	textHandler := slog.NewTextHandler(output, &opts)
	logger := slog.New(textHandler)

	return *logger
}
