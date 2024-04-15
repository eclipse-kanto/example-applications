import time
import json
import paho.mqtt.client as mqtt

# Mock Ditto and Kuksa client interactions
class MockDittoClient:
    def __init__(self):
        self.device_info = None
        self.features = {}

    def send(self, cmd_envelope):
        feature_id, property_path, value = cmd_envelope.split(":")
        self.features[f"{feature_id}/{property_path}"] = value
        print(f"Mock Ditto: Updated feature '{feature_id}' property '{property_path}' with value '{value}'")

class MockKuksaClient:
    def __init__(self):
        self.values = {
            'Altitude': 100,
            'Latitude': 17.3850,
            'Longitude': 78.4667,
            'Speed': 50
        }
        self.subscriptions = {}
        self.update_interval = 1  # Seconds between simulated data updates

    def subscribe(self, topic, callback):
        print(f"Mock Kuksa subscribed to topic: {topic}")
        self.subscriptions[topic] = callback
        callback(self.values[topic])

    def simulate_updates(self):
        while True:
            # Simulate changes in some values
            self.values['Speed'] += 2
            # Update subscribed callbacks
            for topic, callback in self.subscriptions.items():
                callback(self.values[topic])
            time.sleep(self.update_interval)

class EdgeClient:
    def __init__(self, mqtt_host="localhost", mqtt_port=1883, mqtt_username=None, mqtt_password=None):
        self.ditto_client = MockDittoClient()
        self.kuksa_client = MockKuksaClient()
        self.device_info = None

        # MQTT settings
        self.mqtt_host = mqtt_host
        self.mqtt_port = mqtt_port
        self.mqtt_username = mqtt_username
        self.mqtt_password = mqtt_password
        self.mqtt_client = mqtt.Client()

        # Set MQTT username and password if provided
        if mqtt_username and mqtt_password:
            self.mqtt_client.username_pw_set(mqtt_username, mqtt_password)

        # Subscribe to specified MQTT topics
        self.subscribe_to_topics()

    def subscribe_to_topics(self):
        def on_message(client, userdata, message):
            topic = message.topic
            payload = json.loads(message.payload)

            # Update internal state based on received topic and payload
            if topic in self.kuksa_client.values:
                self.kuksa_client.values[topic] = payload
                self.handle_kuksa_update(topic, payload)

            print(f"Received message on topic '{topic}': {payload}")

        self.mqtt_client.on_message = on_message
        self.mqtt_client.connect(self.mqtt_host, self.mqtt_port)
        self.mqtt_client.loop_start()

        # Subscribe to Kuksa data topics
        topics_to_subscribe = [
            "Altitude",
            "Latitude",
            "Longitude",
            "Speed"
        ]
        for topic in topics_to_subscribe:
            self.mqtt_client.subscribe(topic)

    def handle_kuksa_update(self, topic, value):
        # Handle Kuksa data signal (simplified)
        print(f"Received Kuksa data signal - Topic: {topic}, Value: {value}")
        # Trigger Ditto data update
        self.update_ditto_feature(topic, value)

    def update_ditto_feature(self, property_path, value):
        # Update Ditto feature with the new value
        feature_id = "VSS"  # Example feature ID (Vehicle Speed and Location)
        cmd_envelope = f"{feature_id}:{property_path}:{value}"
        self.ditto_client.send(cmd_envelope)

    def start(self):
        try:
            # Simulate device info retrieval
            self.device_info = EdgeDeviceInfo()
            self.device_info.unmarshal_json('{"deviceId": "123", "tenantId": "456", "policyId": "789"}')

            # Add VSS feature to mock Ditto client
            self.add_vss_feature()

            # Simulate Kuksa data updates
            self.kuksa_client.simulate_updates()

        except Exception as e:
            print(f"Error occurred during start: {e}")

    def add_vss_feature(self):
        # Add VSS feature to mock Ditto client (simplified)
        feature_id = "VSS"
        feature_data = {
            "description": "Vehicle Speed and Location",
            "properties": self.kuksa_client.values.copy()  # Include a copy of Kuksa values in feature data
        }
        self.ditto_client.features[feature_id] = feature_data
        print(f"Mock Ditto: Added feature '{feature_id}' with properties")
        print(feature_data)

    def stop(self):
        self.mqtt_client.loop_stop()
        self.mqtt_client.disconnect()

class EdgeDeviceInfo:
    def __init__(self):
        self.deviceId = None
        self.tenantId = None
        self.policyId = None

    def unmarshal_json(self, payload):
        data = json.loads(payload)
        self.deviceId = data.get("deviceId")
        self.tenantId = data.get("tenantId")
        self.policyId = data.get("policyId")

if __name__ == "__main__":
    # Specify MQTT host, port, username, and password
    mqtt_host = "localhost"
    mqtt_port = 1883
    mqtt_username = "Username"
    mqtt_password = "password"

    # Create EdgeClient instance with MQTT credentials
    edge_client = EdgeClient(mqtt_host, mqtt_port, mqtt_username, mqtt_password)
    edge_client.start()  # Start the edge client simulation
