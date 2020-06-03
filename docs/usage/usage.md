# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://pkg.go.dev/github.com/unit-io/unitd) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/unitd)](https://goreportcard.com/report/github.com/unit-io/unitd) [![Coverage Status](https://coveralls.io/repos/github/unit-io/unitd/badge.svg?branch=master)](https://coveralls.io/github/unit-io/unitd?branch=master)

<p align="left">
  <img src="unitd.png" width="300" alt="Unitd" title="Unitd: Blazing Fast and Secure Messaging Broker"> 
</p>

# unitd: Blazing Fast and Secure Messaging Broker

## unitd is an open source messaging broker for IoT and other real-time messaging service. Unitd messaging API is built for speed and security.

Unitd is a real-time messaging service for IoT connected devices, it is based on MQTT protocol. Unitd is blazing fast and secure messaging infrastructure and APIs for IoT, gaming, apps and real-time web.

Unitd can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Unitd is perfect for the internet of things and internet connected devices.

## Quick Start
To build Unitd from source code use go get command and copy [unitd.conf](https://github.com/unit-io/unitd/tree/master/unitd.conf) to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd && unitd

## Usage

Files under examples folder has various examples for unitd usage. Code snippet is given below to use unitd messaging broker with web socket and javascript client.

You need to register a client id and generate keys in order to publish or subscribe to topics.

Add below mqtt client script to the web page

>  <script src="https://cdnjs.cloudflare.com/ajax/libs/paho-mqtt/1.0.1/mqttws31.js" type="text/javascript"></script>

Send empty client Id (as shown in the below code snippet) in order to register a new client. The unitd generate two primary client Ids. These client Ids are used to request secondary client Ids. You could request multiple of secondary client Ids in order to send and receive messages to multiple topics. For isolation of topics use secondary Ids those are generated using distinct primary Ids. 

> <script type="text/javascript">client = new Paho.MQTT.Client("127.0.01", 6060, "");</script>

Next step: send primary client Id in order to request secondary client Id. The code snippet is given below to request client Id.

```
script type="text/javascript">
client = new Paho.MQTT.Client("127.0.01", 6060, "<<primary clientid>>");
// set callback handlers
client.onConnectionLost = onConnectionLost;
client.onMessageArrived = onMessageArrived;

// connect the client
client.connect({onSuccess:onConnect});

// called when the client connects
function onConnect() {
  // Once a connection has been made, send payload to request secondary clientid.
  console.log("onConnect");
   payload = JSON.stringify({"type":"id"});
   message = new Paho.MQTT.Message(payload);
   message.destinationName = "unitd/clientid";
   client.send(message);
}

// called when the client loses its connection
function onConnectionLost(responseObject) {
  if (responseObject.errorCode !== 0) {
    console.log("onConnectionLost:"+responseObject.errorMessage);
  }
}

// called when a message arrives
function onMessageArrived(message) {
  console.log("onMessageArrived:"+message.payloadString);
}
</script>
```

To subscribe to topic and publish messages to a topic generate key for the topic. Key is separated from topic using "/" character.

```
// called when the client connects
function onConnect() {
  // Once a connection has been made, send payload to request key to publish or subscribe a topic. Pass "rw" to the Type field to read and write to the topic.
  console.log("onConnect");
    payload = JSON.stringify({"topic":"unit.b","type":"rw"});
    message = new Paho.MQTT.Message(payload);
    message.destinationName = "unitd/keygen";
    client.send(message);
}
```

To publish and subscribe to the topic use a valid key.

```
// called when the client connects
function onConnect() {
  // Once a connection has been made, publish and subscribe to a topic and use a valid key. Topics are separated by "." character, use * as wildcard character.
  console.log("onConnect");
   client.subscribe("<<key>>/unit.b");
   message = new Paho.MQTT.Message("Hello unit.b!");
   message.destinationName = "<<key>>/unit.b";
   client.send(message);
}
```

Use dot '.' character as topic separator and use three dots '`...`' at the end to subscribe to all topics following the path or use '`*`' character to subscribe to single wildcard topic.

Following are valid topic subscriptions:
- "key/`...`"
- "key/unit.b`...`"
- "key/unit.`*`.b.`*`"

You could send messages to tenant unit.b.b1:
- "key/unit.b.b11"
```
// called when the client connects
function onConnect() {
  // Once a connection has been made, publish and subscribe to a topic and use a valid key.
  console.log("onConnect");
   message = new Paho.MQTT.Message("Hello b11!");
   message.destinationName = "<<key>>/unit.b.b11;
   client.send(message);
}
```

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/unitd/blob/master/LICENSE).
