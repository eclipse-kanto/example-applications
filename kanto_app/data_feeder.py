from kuksa_client.grpc import VSSClient
import json
import paho.mqtt.client as mqtt

# Define MQTT broker settings
MQTT_BROKER_HOST = 'localhost'
MQTT_BROKER_PORT = 1883  # Default MQTT port
# MQTT_TOPIC = 'vehicle/speed'# MQTT topic to publish to
MQTT_PUBLISH = 'test'
MQTT_USERNAME = '<username>'
MQTT_PASSWORD = '<password>'

# MQTT client setup
mqtt_client = mqtt.Client()

# Set MQTT username and password
mqtt_client.username_pw_set(MQTT_USERNAME, MQTT_PASSWORD)

# Connect to MQTT broker
mqtt_client.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT)
mqtt_client.loop_start()  # Start the MQTT client network loop

try:
    with VSSClient('127.0.0.1', 55555) as client:
        for updates in client.subscribe_current_values([
            'Vehicle.Speed',
        ]):
            speed = updates['Vehicle.Speed'].value
            print(f"Received updated speed: {speed}")

            # Prepare message to publish
            message = {'speed': speed}

            # Publish message to MQTT broker
            mqtt_client.publish(MQTT_PUBLISH, json.dumps(message))

except KeyboardInterrupt:
    pass

finally:
    mqtt_client.disconnect()  # Disconnect MQTT client
    mqtt_client.loop_stop()  # Stop the MQTT client network loop
