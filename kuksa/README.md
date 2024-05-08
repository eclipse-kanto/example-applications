![Kanto logo](https://github.com/eclipse-kanto/kanto/raw/main/logo/kanto.svg)

# Eclipse Kanto - Eclipse Kuksa Integration

# Introduction

This is an example application that connects to [Eclipse Kuksa Databroker](https://github.com/eclipse-kuksa/kuksa-databroker) and demonstrates how the COVESA Vehicle Signal Specification(VSS) data could be transformed an AWS IoT Shadow using Eclipse Kanto.

# Installation

## Prerequisites
You must have an installed and working instance of:
* Eclipse Kanto Container Management
* Eclipse Kanto AWS Connector that is connected to a Thing in AWS IoT Core
* [Python 3](https://wiki.python.org/moin/BeginnersGuide/Download) and [pip3](https://pip.pypa.io/en/stable/installation/#)

## Steps
Create and start Kuksa Databroker
```shell
kanto-cm create --network=host --name=server ghcr.io/eclipse-kuksa/kuksa-databroker:0.4.4 --insecure
kanto-cm start --name=server
```

Create and start Kuksa Databroker CLI in a dedicated terminal, where VSS data will be fed at later point
```shell
kanto-cm create --i --t --network=host --name=client ghcr.io/eclipse-kuksa/kuksa-databroker-cli:0.4.4 --server localhost:55555
kanto-cm start --i --a --name=client
```

Install required Python dependencies and run the script
```shell
pip3 install -r requirements.txt
python3 ./edge_client.py 
```

There should be a new device shadow names 'VSS' in your AWS IoT Thing. With the VSS data from the Kuksa Databroker server displayed as a shadow state
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

If you go back to the Kuksa Databroker CLI terminal, you can feed new data to the Kuksa Databroker server
```shell
feed Vehicle.Speed 120
feed Vehicle.CurrentLocation.Altitude 640
feed Vehicle.CurrentLocation.Latitude 43
feed Vehicle.CurrentLocation.Longitude 25
```

The python script is subscribed for changes of this VSS data paths and the values are updated in the VSS shadow as well
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
The python scripts allows configuring of connection settings for the local MQTT broker and the Kuksa Databroker and also which VSS date paths to be followed. The allowed arguments and their default values could be listed with
```shell
python3 ./edge_client.py --help
```