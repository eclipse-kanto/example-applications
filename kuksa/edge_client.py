# Copyright (c) 2024 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Eclipse Public License 2.0 which is available at
# https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
# which is available at https://www.apache.org/licenses/LICENSE-2.0.
#
# SPDX-License-IDentifier: EPL-2.0 OR Apache-2.0

import argparse
import signal
import sys

import paho.mqtt.client as mqtt
from ditto.client import Client
from ditto.model.feature import Feature
from ditto.protocol.things.commands import Command
from kuksa_client import KuksaClientThread

from edge_device_info import EdgeDeviceInfo
from utils import process_tree, process_signal

FEATURE_ID_VSS = "VSS"

EDGE_CLOUD_CONNECTOR_TOPIC_DEV_INFO = "edge/thing/response"
EDGE_CLOUD_CONNECTOR_TOPIC_DEV_INFO_REQUEST = "edge/thing/request"

DATA_PATHS = "Vehicle.CurrentLocation.Altitude,Vehicle.CurrentLocation.Latitude,Vehicle.CurrentLocation.Longitude,Vehicle.Speed"

class EdgeClient:
    def __init__(self, host, port, paths):
        self.mqtt_client = None
        self.kuksa_client = KuksaClientThread(config={'ip':host,'protocol': 'grpc', 'port': port, 'insecure': True})
        self.device_info = None
        self.ditto_client = None
        self.data_paths = paths

    def on_connect(self, client:mqtt.Client, obj, flags, rc):
        print("Connected with result code:", str(rc))
        self.mqtt_client = client
        # init ditto client
        self.ditto_client = Client(paho_client=self.mqtt_client)
        self.ditto_client.connect()
        # init kuksa client
        self.kuksa_client.start()
        # trigger initialization
        self.mqtt_client.subscribe(EDGE_CLOUD_CONNECTOR_TOPIC_DEV_INFO)
        self.mqtt_client.publish(EDGE_CLOUD_CONNECTOR_TOPIC_DEV_INFO_REQUEST, None, 1)

    def on_message(self, client, userdata, msg):
        try:
            if msg.topic == EDGE_CLOUD_CONNECTOR_TOPIC_DEV_INFO:
                if self.device_info is None:
                    self.device_info = EdgeDeviceInfo()
                    self.device_info.unmarshal_json(msg.payload)
                    self.add_vss_feature()
                    self.subscribe()
                else:
                    print('Device info already available - discarding message')
                return
        except Exception as ex:
            print(ex)

    def subscribe(self):
        print('Subscribing to VSS data paths:', self.data_paths)
        self.kuksa_client.subscribeMultiple(self.data_paths, self.on_kuksa_signal)

    def add_vss_feature(self):
        # add the vss feature
        feature = Feature
        cmd = Command(self.device_info.deviceId).feature(FEATURE_ID_VSS).modify(Feature().to_ditto_dict())
        cmd_envelope = cmd.envelope(response_required=False, content_type="application/json")
        self.ditto_client.send(cmd_envelope)

        # add the vss tree as properties
        vss_tree = self.kuksa_client.getValue('Vehicle.*')
        processed = process_tree(vss_tree)
        for key, val in processed.items():
            cmd = Command(self.device_info.deviceId).feature_property(FEATURE_ID_VSS, key.replace('.','/')).modify(val)
            cmd_envelope = cmd.envelope(response_required=False, content_type="application/json")
            self.ditto_client.send(cmd_envelope)

    def on_kuksa_signal(self, message):
        print('Received signal:', message)
        if self.device_info is None:
            print('No device info is initialized to process VSS data')
            return
        processed = process_signal(message)
        # update property
        print('Updating VSS properties:', processed)
        for key, val in processed.items():
            cmd = Command(self.device_info.deviceId).feature_property(FEATURE_ID_VSS, key.replace('.','/')).modify(val)
            cmd_envelope = cmd.envelope(response_required=False, content_type="application/json")
            self.ditto_client.send(cmd_envelope)

    def shutdown(self):
        self.kuksa_client.stop()
        self.ditto_client.disconnect()

def parse_args():
    parser = argparse.ArgumentParser(description="Edge Client with configurable MQTT and Kuksa settings", formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument("--mqtt_host", type=str, default="localhost", help="MQTT broker host")
    parser.add_argument("--mqtt_port", type=int, default=1883, help="MQTT broker port")
    parser.add_argument("--mqtt_username", type=str, default=None, help="MQTT username")
    parser.add_argument("--mqtt_password", type=str, default=None, help="MQTT password")
    parser.add_argument("--kuksa_host", type=str, default="localhost", help="Kuksa Databroker host")
    parser.add_argument("--kuksa_port", type=int, default=55555, help="Kuksa Databroker port")
    parser.add_argument("--paths", type=str, default=DATA_PATHS, help="Comma separated VSS data paths to subscribe to")
    return parser.parse_args()

if __name__ == "__main__":
    args = parse_args()
    
    # Set VSS data paths to subscribe to
    args.paths = [s.strip() for s in args.paths.split(",")]

    paho_client = mqtt.Client()
    edge_client = EdgeClient(args.kuksa_host, args.kuksa_port, args.paths)
    
    # Set MQTT username and password if provided
    if args.mqtt_username and args.mqtt_password:
        self.mqtt_client.username_pw_set(mqtt_username, mqtt_password)

    paho_client.on_connect = edge_client.on_connect
    paho_client.on_message = edge_client.on_message
    paho_client.connect(args.mqtt_host, args.mqtt_port)


    def termination_signal_received(signal_number, frame):
        print("Received termination signal. Shutting down")
        edge_client.shutdown()
        paho_client.disconnect()


    signal.signal(signal.SIGINT, termination_signal_received)
    signal.signal(signal.SIGQUIT, termination_signal_received)
    signal.signal(signal.SIGTERM, termination_signal_received)
    print('before loop forever')
    paho_client.loop_forever()