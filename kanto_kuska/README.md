This code provides a mock implementation of an edge client that interacts with Ditto and Kuksa services. It simulates data flow and feature management.

Functionality:

    Connects to an MQTT broker (configurable)
    Subscribes to topics for Kuksa data (Altitude, Latitude, Longitude, Speed)
    Simulates updates for Kuksa data at specified intervals
    Upon receiving Kuksa data updates:
        Updates internal state
        Triggers Ditto feature update with the new value

Mock Clients:

    MockDittoClient: Simulates interaction with Ditto, updates features based on received data.
    MockKuksaClient: Simulates interaction with Kuksa, provides sensor data and updates.

EdgeClient:

    Establishes connection to MQTT broker with provided credentials.
    Subscribes to relevant Kuksa data topics.
    Processes incoming MQTT messages and updates internal state.
    Triggers Ditto feature updates based on received Kuksa data.
    Simulates device information retrieval.
    Manages Ditto features (adding and updating).

Running the Simulation:

    Configure MQTT settings (host, port, username, password) in the code.
    Run the script: python edge_client.py
