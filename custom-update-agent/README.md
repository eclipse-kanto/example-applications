![Kanto logo](https://github.com/eclipse-kanto/kanto/raw/main/logo/kanto.svg)

# Eclipse Kanto - Files Update Agent

# Introduction

This is an example application for demonstrating how an update agent works in implementing the Update Agent API to update a certain domain inside the target device. In the case of the Files Update Agent application it is used to manage the files inside of a directory on the target device.
 
# Desired state

The update process is initiated by sending the desired state specification as an MQTT message towards the device, which is handled by the Update Manager component.

The desired state specification in the scope of the Update Manager is a JSON-based document, which consists of multiple component definitions per domain. The Files Update Agent is responsible for the files domain. Below is an example desired state that contains the Eclipse Kanto logo.

```json
{
  "desiredState": {
    "domains": [
      {
        "id": "files",
        "config": [],
        "components": [
          {
            "config": [
              {
                "key": "file_name",
                "value": "kantoLogo.svg"
              },
              {
                "key": "download_url",
                "value": "https://github.com/eclipse-kanto/kanto/raw/main/logo/kanto.svg"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

# Commands

Based on the received desired state the update agent can do the following changes to the provided directory:

- Download file
- Remove file
- Replace file

# Installation

## Prerequisites
You must have an installed and working instance of:
* Eclipse Kanto Update Manager
* Eclipse Kanto Container Management


Regardless if you are running the update agent as a standard or containerized application, you will need to add the following to the update manager configuration, located at `/etc/update-manager/config.json`:

```json
{
  "domain": "device",
  "agents": {
    "files": {
        "rebootRequired": false,
        "readTimeout": "20s"
    }
  }
}
```
After that reboot the service by executing:

```
$ sudo systemctl restart kanto-update-manager.service
```
## Standard service
Replace the directory provided with `-dir` flag in `custom-update-agent.service` with the desired file directory.
``` Ini
[Unit]
Description=Eclipse Kanto - Files Update Agent
[Service]
Type=simple
ExecStart=/usr/bin/custom-update-agent --dir "<files-directory>"
Restart=always
TimeoutSec=300

[Install]
WantedBy=multi-user.target
```
After that execute the provided `install.sh` script:
```
$ sudo ./install.sh
``` 
To check the status of the `custom-update-agent.service` execute:
```
$ systemctl status custom-update-agent.service
```
## Containerized application
A containerized instance of the Files Update Agent can be built using Eclipse Kanto Container Management. 
Replace the fields marked with <> in `deployment.json` accordingly
```json
{
    "container_name": "files-update-agent",
    "image": {
        "name": "<host>:<port>/<image>:<version>"
    },
    "host_config": {
      "network_mode": "host"
    },
    "mount_points": [
        {
            "destination": "<certificates-directory>",
            "source": "<certificates-directory>",
            "propagation_mode": "rprivate"
        },
        {
            "destination": "/bin/fileagent",
            "source": "<files-directory>",
            "propagation_mode": "shared"
        }
    ]
}
```
Create the container by executing:
```
$ sudo kanto-cm create -f deployment.json
```
You can check the status of the container by executing:
```
$ sudo kanto-cm list
```
Start the containerized application by executing:
```
$ sudo kanto-cm start <container-id>
```
You can check the logs of the container by executing:
```
$ sudo kanto-cm logs --debug <container-id>
```