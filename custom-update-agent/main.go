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

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eclipse-kanto/example-applications/custom-update-agent/updateagent"
	"github.com/eclipse-kanto/example-applications/custom-update-agent/util"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/mqtt"
)

func main() {
	logger := util.ConfigLogger(slog.LevelDebug, os.Stdout)
	slog.SetDefault(&logger)

	fileDirPtr := flag.String("dir", "./fileagent", "the path to the directory where file agent will manage files")
	flag.Parse()
	updateagent.FileDirectory = *fileDirPtr

	updateAgent, _ := updateagent.Init(*mqtt.NewDefaultConfig(), "files")
	err := updateAgent.(api.UpdateAgent).Start(context.Background())
	if err != nil {
		slog.Error("could not start Update Agent service! got", "error", err)
	} else {
		slog.Info("successfully started Update Agent service")
	}

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-signalChan
	slog.Info("Exiting!, recieved", "signal", sig)
	updateAgent.(api.UpdateAgent).Stop()

}
