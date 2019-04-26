# Trace

<p align="left">
  <img src="trace.png" width="70" alt="Trace" title="Trace: Blazing Fast and Secure Messaging Broker"> 
</p>

# Trace: Blazing Fast and Secure Messaging Broker

## Trace is an open source messaging borker for IoT and other real-time messaging service. Trace messaging API is built for speed and security.

Trace is a real-time messaging service for IoT connected devices, it is based on MQTT protocol. Trace is blazing fast and secure messaging infrastructure and APIs for IoT, gaming, apps and real-time web.

Trace can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Trace is perfect for the internet of things and internet connected devices.

## Quick Start
To build trace from source code use go get command.

> go get -u github.com/tracedb/trace && trace

## Usage

Files under web folder has various usage examples. Code snippet is given below to use trace messaging broker with web socket and javascript client.

You need to register a client id and request key in order to publish or subscribe to a topic.

Add below mqtt client script to the web page

>  <script src="https://cdnjs.cloudflare.com/ajax/libs/paho-mqtt/1.0.1/mqttws31.js" type="text/javascript"></script>

Send empty client Id (as shown in the below code snippet) in order to register a new client. Trace generate two primary client ids. These client Ids are used to request secondary client Ids. You could request n-number of secodary client Ids in order to send and receive messages to multiple topics. For isolation of messaging between topics use secondary Ids those are generated using distinct primary Ids. 

> <script type="text/javascript">client = new Paho.MQTT.Client("127.0.01", 6060, "");</script>

Next step: send primary client Id in order to request secondary client Id.

```
script type="text/javascript">
client = new Paho.MQTT.Client("127.0.01", 6060, "UBUWKLGQFNFARADMdOHNccSDSccXTYQcIfKOSYGJQBIaQeFWIabA");
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
   message.destinationName = "trace/clientid";
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

Now use secondary client Id to publish and subscribe messages. To publish and subscribe messages, you need a key prefix to a topic. Key is separated from topic using "/" character.

```
// called when the client connects
function onConnect() {
  // Once a connection has been made, send payload to request key to publish or subscribe a topic. Pass "rw" to the Type field to read and write to the topic.
  console.log("onConnect");
    payload = JSON.stringify({"topic":"dev18.world","type":"rw"});
    message = new Paho.MQTT.Message(payload);
    message.destinationName = "tace/keygen";
    client.send(message);
}
```

Now you can publish and subscribe to the topic using a valid key.
```
// called when the client connects
function onConnect() {
  // Once a connection has been made, publish and subscribe to a topic and use a valid key. Topics are separated by "." character, use * as wildcard character.
  console.log("onConnect");
   client.subscribe("EbQCNYBTTSJFO/dev18.world");
   message = new Paho.MQTT.Message("Hello World!");
   message.destinationName = "EbQCNYBTTSJFO/dev18.world";
   client.send(message);
}
```

Use dot '.' character as topic separator and use three dots '...' at the end to subscribe to all topics following the path. Use '*' character to subscibe to single wildcard topic.

Following are valid topic subsriptions:
- "EYACMAAVDOZKC/..."
- "EZQCMYAUEFbLe/dev18.world..."
- "EZISMUJQBPAWK/dev18.&ast;.usa.&ast;.&ast;.&ast;.empire-state.&ast;"

You could now send messages to empire-state tenants:
- "EbfaNfbYdFTMI/dev18.world.usa.newyork.nyc.manhattan.empire-state.tenant1"
```
// called when the client connects
function onConnect() {
  // Once a connection has been made, publish and subscribe to a topic and use a valid key.
  console.log("onConnect");
   message = new Paho.MQTT.Message("Hello Tenant1!");
   message.destinationName = "EbfaNfbYdFTMI/dev18.world.usa.newyork.nyc.manhattan.empire-state.tenant1;
   client.send(message);
}
```

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2019 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/tracedb/trace/blob/master/LICENSE).
