![Kanto logo](https://github.com/eclipse-kanto/kanto/raw/main/logo/kanto.svg)

# Eclipse Kanto - Eclipse Kuksa Integration

# Introduction

This is an example application that connects to [Eclipse Kuksa Databroker](https://github.com/eclipse-kuksa/kuksa-databroker) and demonstrates how the COVESA Vehicle Signal Specification(VSS) data could be transformed to a digital twin using Eclipse Kanto. For the application it is transparent which Kanto cloud connectivity options is used(AWS, Azure or Suite). Nevertheless, for the completeness of the next guide steps, an AWS connector is chosen, so the VSS data from a Kuksa Databroker will be presented as an AWS IoT Shadow.

# Installation

## Prerequisites
You must have an installed and working instance of:
* Eclipse Kanto Container Management
* Eclipse Kanto AWS Connector that is connected to a Thing in AWS IoT Core

## Steps
Create  container and start the Kuksa Databroker
```shell
kanto-cm create --name=databroker ghcr.io/eclipse-kuksa/kuksa-databroker:0.4.4 --insecure
kanto-cm start --name=databroker
```

Create container and start the Kuksa Databroker CLI in a dedicated terminal, where VSS data will be fed at later point
```shell
kanto-cm create --name=cli --i --t --hosts=databroker:container_databroker-host --rp=no ghcr.io/eclipse-kuksa/kuksa-databroker-cli:0.4.4  --server=databroker:55555
kanto-cm start --i --a --name=cli
```

Create container and start the Kuksa Example Application
```shell
kanto-cm create -f ./deployment.json
kanto-cm start --name=vss
```

There should be a new device shadow named 'VSS' in your AWS IoT Thing. With the VSS data from the Kuksa Databroker displayed as a shadow state
```json
{
  "state": {
    "reported": {
      "Vehicle": {
        "Length": {
          "timestamp": "2024-05-06T12:31:33.732487+00:00",
          "value": 0
        },
        "CurbWeight": {
          "timestamp": "2024-05-06T12:31:33.732469+00:00",
          "value": 0
        },
        "GrossWeight": {
          "timestamp": "2024-05-06T12:31:33.732484+00:00",
          "value": 0
        },
        "StartTime": {
          "timestamp": "2024-05-06T12:31:33.732633+00:00",
          "value": "0000-01-01T00:00Z"
        },
        "MaxTowWeight": {
          "timestamp": "2024-05-06T12:31:33.732493+00:00",
          "value": 0
        },
        "MaxTowBallWeight": {
          "timestamp": "2024-05-06T12:31:33.732492+00:00",
          "value": 0
        },
        "Speed": {
          "timestamp": "2024-05-06T12:31:33.732486+00:00",
          "value": 0
        },
        "Height": {
          "timestamp": "2024-05-06T12:31:33.732485+00:00",
          "value": 0
        },
        "Width": {
          "timestamp": "2024-05-06T12:31:33.732869+00:00",
          "value": 0
        }
      }
    }
  }
}
```

You can go back to the Kuksa Databroker CLI terminal and feed new data to the Kuksa Databroker
```shell
feed Vehicle.Speed 120
feed Vehicle.CurrentLocation.Altitude 640
feed Vehicle.CurrentLocation.Latitude 43
feed Vehicle.CurrentLocation.Longitude 25
```

The Kuksa Example Application is subscribed for changes of this VSS data paths and the values are updated in the VSS shadow as well
```json
{
  "state": {
    "reported": {
      "Vehicle": {
        ...
        "Speed": {
          "timestamp": "2024-05-06T15:11:12.911755+00:00",
          "value": 120
        },
        ...
        "CurrentLocation": {
          "Altitude": 640,
          "Latitude": 43,
          "Longitude": 25
        }
      }
    }
  }
}
```

# Control
The Kuksa Example Application is based on python scripts that allows configuring of connection settings for the local MQTT broker and the Kuksa Databroker and also which VSS data paths to be followed. Allowed arguments and their default values:
| Argument     | Type | Default |     Description     |
| --------     | ---- | ------- |     -----------     |
|mqtt_host     |string|localhost|MQTT broker host     |
|mqtt_port     |int   |1883     |MQTT broker port     |
|mqtt_username |string|         |MQTT username        |
|mqtt_password |string|         |MQTT password        |
|kuksa_host    |string|localhost|Kuksa Databroker host|
|kuksa_port    |int   |55555    |Kuksa Databroker port|
|vss_paths     |string|Vehicle.CurrentLocation.Altitude,Vehicle.CurrentLocation.Latitude,Vehicle.CurrentLocation.Longitude,Vehicle.Speed| Comma separated VSS data paths to subscribe to |
|log_level     |string| info    |Logging level, possible values are critical,fatal,error,warn,warning,info,debug,notset