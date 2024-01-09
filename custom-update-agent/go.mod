module github.com/eclipse-kanto/example-applications/custom-update-agent

go 1.18

replace github.com/docker/docker => github.com/moby/moby v23.0.3+incompatible

require (
	github.com/eclipse-kanto/container-management v0.1.0-M4
	github.com/eclipse-kanto/update-manager v0.0.0-20230628072101-b91f0c30e00f
	github.com/pkg/errors v0.9.1
	github.com/rickar/props v1.0.0
)

require (
	github.com/eclipse/paho.mqtt.golang v1.4.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
