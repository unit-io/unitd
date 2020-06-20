# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://godoc.org/github.com/unit-io/unitd) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/unitd)](https://goreportcard.com/report/github.com/unit-io/unitd) [![Coverage Status](https://coveralls.io/repos/github/unit-io/unitd/badge.svg?branch=master)](https://coveralls.io/github/unit-io/unitd?branch=master)

## Unitd is an open source messaging broker for IoT and other real-time messaging service. Unitd messaging API is built for speed and security.

Unitd is a real-time messaging service for IoT connected devices, it is based on MQTT protocol. Unitd is blazing fast and secure messaging infrastructure and APIs for IoT, gaming, apps and real-time web.

Unitd can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Unitd is perfect for the internet of things and internet connected devices.

## Quick Start
To build [unitd](https://github.com/unit-io/unitd) from source code use go get command and copy unitd.conf to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd && unitd

## Unitd Clients
- [unitd-go](https://github.com/unit-io/unitd-go) Go client to pubsub messages over protobuf using GRPC application
- [unitd-ws](https://github.com/unit-io/unitd-ws) Javascript client to pubsub messages over websocket using MQTT protocol. 

## Quick Start
To build Unitd from source code use go get command and copy [unitd.conf](https://github.com/unit-io/unitd/tree/master/unitd.conf) to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd && unitd

## Usage
The examples folder has various examples for unitd usage. Code snippet is given below to use unitd messaging broker with web socket and javascript client.

First you need to register a client id. To get new client id send connect request using blank client id. The unitd generates two primary client Ids.

Add below mqtt client script to the web page

>  <script src="https://cdnjs.cloudflare.com/ajax/libs/paho-mqtt/1.0.1/mqttws31.js" type="text/javascript"></script>
>  <script src="../unitd.js" type="text/javascript"></script>

Send empty client Id (as shown in the below code snippet) in order to register a new client. The unitd generate two primary client Ids. These client Ids are used to request secondary client Ids. You could request multiple of secondary client Ids in order to send and receive messages to multiple topics. For isolation of topics use secondary Ids those are generated using distinct primary Ids. 

> <script type="text/javascript">client = new Paho.MQTT.Client("127.0.01", 6060, "");</script>
```

    // Initialize new Paho client connection
    client = new Paho.MQTT.Client("127.0.01", 6060, "");

    // Set callback handlers
    client.onConnectionLost = onConnectionLost;
    client.onMessageArrived = onMessageArrived;

    // Connect the client, if successful, call onConnect function
    client.connect({
        onSuccess: onConnect,
    });

```

Next step: send primary client Id in order to request secondary client Id. The code snippet is given below to request client Id. Cclient id request type is 0. You cannot request type 1 i.e. primary Id but you must use primary client Id for the connection in order to make request of secondary client Ids. You can request multiple secondary client Ids.

Note, for topic isolation use client Ids generated from distinct primary client Ids.

```

    client = new Paho.MQTT.Client("127.0.01", 6060, "<<primary clientid>>");
    payload = JSON.stringify({"type":"0"});
    message = new Paho.MQTT.Message(payload);
    message.destinationName = "unitd/clientid";
    client.send(message);

```

To subscribe to topic and publish messages to a topic generate key for the topic.

```
    // Once a connection has been made, send payload to request key to publish or subscribe a topic. Pass "rw" to the Type field to set read/write permissions for key on topic.
    payload = JSON.stringify({"topic":"teams.alpha.ch1","type":"rw"});
    message = new Paho.MQTT.Message(payload);
    message.destinationName = "unitd/keygen";
    client.send(message);

```

To publish and subscribe to the topic use a valid key.  Key is separated from topic using "/" character.

```
    // Once a connection has been made, publish and subscribe to a topic and use a valid key. Topics are separated by "." character, use * as wildcard character.
    // Subscribe to team alpha channe1.
    client.subscribe("<<key>>/teams.alpha.ch1");
    // Publish message to team alpha channel1.
    message = new Paho.MQTT.Message("Message for teal alpha channel1!");
    message.destinationName = "<<key>>/teams.alpha.ch1";
    client.send(message);

```

Use dot '.' character as topic separator and use three dots '`...`' at the end to subscribe to all topics following the path or use '`*`' character to subscribe to single wildcard topic.

Following are valid topic subscriptions:
Subscribe to team alpha all channels
- "key/teams.alpha`...`"

Subscribe to channel1 of any team
- "key/teams.`*`.ch1"

Send messages to channel and channel receivers:

```
    // Once a connection has been made, publish and subscribe to a topic and use a valid key.
    message = new Paho.MQTT.Message("Hello team alpha channel1!");
    message.destinationName = "<<key>>/teams.alpha.ch1;
    client.send(message);

    message = new Paho.MQTT.Message("Hello team alpha channel1 receiver1!");
    message.destinationName = "<<key>>/teams.alpha.ch1.u1;
    client.send(message);

```

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/unitd/blob/master/LICENSE).
