# Edge Client Simulation with MQTT

This project simulates an edge client using MQTT for communication, with mock devices (Ditto and Kuksa clients) generating and receiving data.

## Prerequisites

- Docker installed on your machine

## How to Run

### 1. Build the Docker Image

Navigate to the directory containing the Dockerfile and Python script:

```bash
docker build -t edge-client .


Run the Docker Container

Replace <mqtt_host>, <mqtt_port>, <mqtt_username>, and <mqtt_password> with your MQTT broker configuration.

docker run --rm edge-client --host <mqtt_host> --port <mqtt_port> --username <mqtt_username> --password <mqtt_password>


Sample Output

Once the Docker container is running, you should see output similar to the following in your terminal:

Mock Ditto: Added feature 'VSS' with properties
{'description': 'Vehicle Speed and Location', 'properties': {'Altitude': 100, 'Latitude': 17.385, 'Longitude': 78.4667, 'Speed': 50}}
Received Kuksa data signal - Topic: Speed, Value: 52
Mock Ditto: Updated feature 'VSS' property 'Speed' with value '52'
Received message on topic 'Speed': 52
Received Kuksa data signal - Topic: Speed, Value: 54
Mock Ditto: Updated feature 'VSS' property 'Speed' with value '54'
Received message on topic 'Speed': 54
...
