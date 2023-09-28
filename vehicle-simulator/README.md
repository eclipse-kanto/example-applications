![Kanto logo](https://github.com/eclipse-kanto/kanto/raw/main/logo/kanto.svg)

# Eclipse Kanto - Vehicle Simulator

# Introduction

The simulator is used to feed simulated vehicle data and showcase the cloud connectivity that comes with Kanto. The data is really simple and simulates a small portion of actual data that can be sent by actual vehicle. This data is sent to the local mqtt broker and depending on the Kanto configuration uploaded in the corresponding cloud.
# Telemetry data

The simulator sends periodic updates for a ditto feature called *OBD*. This feature contains information for the Control Module Voltage and for three standard DTCs (Device Trouble Code) b1601, b1602 and b1603. The values are randomly generated. Here is a sample ditto message.

```json
{
   "topic":"<YourDeviceId>:Vehicle/things/twin/commands/modify",
   "headers":{
      "content-type":"application/json",
      "response-required":false
   },
   "path":"/features/OBD",
   "value":{
      "properties":{
         "controlModuleVoltage": 9,
         "b1601": true,
         "b1602": false,
         "b1603": true
      }
   }
}
```

Another data sent from the simulator but only after some user interaction, which will be explained later is a ditto feature called *trunk*. It contains information if the trunk is open or closed. A sample message for the trunk state looks like this.

```json
{
   "topic":"<YourDeviceId>:Vehicle/things/twin/commands/modify",
   "headers":{
      "content-type":"application/json",
      "response-required":false
   },
   "path":"/features/trunk",
   "value":{
      "properties":{
         "isOpen":true
      }
   }
}
```

# Commands

The simulator supports the following commands.

- Open Trunk
- Reset DTC

## Open Trunk

When Open Trunk command is received, the simulator sends a *trunk* feature update, reporting *"isOpen" : true*, simulating opening of the trunk. After predefined, but configurable amount of time, the simulator sends *trunk* feature update, reporting *"isOpen" : false*, simulating closing of the trunk.

To invoke the command one needs to send a message from the cloud instance to the corresponding connected device, where the simulator is running. The exact way to do this depends on the cloud itself, but the message payload has to be in the following ditto format.

```json
{
  "topic": "<YourDeviceId>:Vehicle/things/twin/commands/openTrunk",
  "headers": {
  "content-type": "application/json",
    "response-required": false
  },
  "path": "/features/trunk/inbox/messages/openTrunk"
}
```

## Reset DTC

When Reset DTC command is received, the simulator stops the random generation of the three DTCs (b1601, b1602, b1603). They are still sent periodically while updating the *OBD* feature, including the *controlModuleVoltage*, but with constant *false* value, simulating normal working of the vehicle and no trouble codes being raised. After the predefined timeout (the same as for the open trunk command) expires the simulator returns to the regular working mode and starts generating random boolean values for the three DTCs.

To invoke the command one needs to send a message from the cloud instance to the corresponding connected device, where the simulator is running. The exact way to do this depends on the cloud itself, but the message payload has to be in the following ditto format.

```json
{
  "topic": "<YourDeviceId>:Vehicle/things/twin/commands/resetDTC",
  "headers": {
  "content-type": "application/json",
    "response-required": false
  },
  "path": "/features/DTC/inbox/messages/resetDTC"
}
```

# Exponential back-off

The periodic telemetry update supports exponential back-off to avoid generating huge amount of messages if accidentally left unattended. After a predefined, but configurable amount of time, the system enters *Idle State* and the period of the periodic updates is doubled after each execution. The periodic update period is reset to the initial one after executing some of the commands.

# Control

The simulator allows configuring the timeouts discussed above. This can be done only before starting up the simulator.

The executable file which is used to start the simulator provides help information providing a brief reference for the available options. Here is an extended description of the available options:

- ctd-effect-delay - Controls how long the the commands sent from the cloud will be effective. Expected value is integer number representing time period in seconds. The default is 30 seconds. Changing this argument will allow you to have the trunk open and no active DTCs for the specified amount of seconds, after executing the corresponding command.
- idle-state-time  - Determines the delay in seconds to wait before going into *Idle State*. Expected value is integer number. The default is 120 seconds. After entering *Idle State* the simulator doubles the periodic updates period after each execution. Use this to ensure regular periodic updates in meaningful time interval.
- periodic-updates-delay - Used to configure periodic updates period. The expected value is integer representing number of seconds. The default value is 5. The minimal allowed value is 5.







