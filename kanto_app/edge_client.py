import json
import paho.mqtt.client as mqtt
from kuksa_client.grpc import VSSClient
from ditto.client import Client
from ditto.model.feature import Feature
from ditto.model.namespaced_id import NamespacedID
from ditto.protocol.things.commands import Command

# Configuration constants
MQTT_HOST = "localhost"
MQTT_PORT = 1883
MQTT_USERNAME = "<username>"
MQTT_PASSWORD = "<password>"
MQTT_PUBLISH_TOPIC = "test/data"
THING_NAMESPACE = "test.ns"
THING_NAME = "test-name"
FEATURE_ID = "MyFeature"

def connect_mqtt_client():
    """Initialize and connect the MQTT client."""
    mqtt_client = mqtt.Client()
    mqtt_client.username_pw_set(MQTT_USERNAME, MQTT_PASSWORD)
    mqtt_client.connect(MQTT_HOST, MQTT_PORT)
    mqtt_client.loop_start()
    return mqtt_client

def create_ditto_client():
    """Initialize and return a Ditto client instance."""
    return Client()

def handle_vehicle_speed_updates():
    """Subscribe to vehicle speed updates and publish commands to Ditto."""
    ditto_client = create_ditto_client()
    mqtt_client = connect_mqtt_client()

    try:
        with VSSClient('127.0.0.1', 55555) as client:
            for updates in client.subscribe_current_values(['Vehicle.Speed']):
                speed = updates['Vehicle.Speed'].value
                print(f"Received updated speed: {speed}")

                # Create a new feature instance with the vehicle speed as property
                feature_to_add = Feature().with_properties(speed=speed)

                # Define the thing ID and feature ID in Ditto
                thing_id = NamespacedID().from_string(f"{THING_NAMESPACE}:{THING_NAME}")

                # Create a command to modify the feature in Ditto
                command = Command(thing_id).feature(FEATURE_ID).modify(feature_to_add.to_ditto_dict())

                # Send the command via MQTT using Ditto client
                envelope = command.envelope(correlation_id="speed-correlation-id", response_required=False)
                ditto_client.send(envelope)
                mqtt_client.publish(MQTT_PUBLISH_TOPIC, json.dumps(envelope.to_ditto_dict()))

    except KeyboardInterrupt:
        pass
    finally:
        mqtt_client.disconnect()
        mqtt_client.loop_stop()
        ditto_client.disconnect()

if __name__ == "__main__":
    handle_vehicle_speed_updates()
