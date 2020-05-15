# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://pkg.go.dev/github.com/unit-io/unitd)

<p align="left">
  <img src="unitd.png" width="300" alt="Unitd" title="Unitd: Blazing Fast and Secure Messaging Broker"> 
</p>

# unitd: Blazing Fast and Secure Messaging Broker

## unitd is an open source messaging broker for IoT and other real-time messaging service. Unitd messaging API is built for speed and security.

Unitd is a real-time messaging service for IoT connected devices, it is based on MQTT protocol. Unitd is blazing fast and secure messaging infrastructure and APIs for IoT, gaming, apps and real-time web.

Unitd can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Unitd is perfect for the internet of things and internet connected devices.

## Quick Start
To build Unitd from source code use go get command.

> go get -u github.com/unit-io/unitd && unitd

## Usage
The unitd supports publish/subscribe to topic. You need to register a client id to connect to the unitd broker and generate keys for topic to publish or subscribe to topics. See [usage guide](https://github.com/unit-io/unitd/tree/master/docs/usage/usage.md). 

## Example Web Application
Open [unitd.html](https://github.com/unit-io/unitd/blob/master/examples/html/unitd.html) under example/html folder in a browser.

## Steps
- Generate Client ID
- Specify new client ID and connect to client
- Specify topics to subscribe/publish messages and generate key
- Specify key to the topics with separator '/' and subscribe to topic
- Specify message to send and publish to topic

### First Client
<p align="left">
  <img src="docs/img/client1.png" /> 
</p>

### Second Client
<p align="left">
  <img src="docs/img/client2.png" /> 
</p>

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2019 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/unitd/blob/master/LICENSE).
