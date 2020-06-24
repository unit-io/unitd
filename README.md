# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://pkg.go.dev/github.com/unit-io/unitd)

## Unitd is an open source messaging broker for IoT and other real-time messaging service. Unitd messaging API is built for speed and security.

Unitd is blazing fast and secure messaging infrastructure and APIs for IoT, gaming, apps and real-time web.

Unitd can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Unitd is perfect for the internet of things and internet connected devices.

# About unitdb 

## Key characteristics
- 100% Go
- Optimized for fast pubsub
- Supports message encryption
- Supports time-to-live on message
- Supports subscribing to wildcard topics

## Unitd Clients
- [unitd-go](https://github.com/unit-io/unitd-go) Go client to pubsub messages over protobuf using GRPC application
- [unitd-ws](https://github.com/unit-io/unitd-ws) Javascript client to pubsub messages over websocket using MQTT protocol.

## Tutorials and Videos
The following screen cast video demonstrate the use of [unitd-ws](https://github.com/unit-io/unitd-ws) javascript client to pubsub messages over websocket using unitd messaging server.

[![Unitd Pubsub](https://img.youtube.com/vi/Wz1F9FQzb_c/0.jpg)](https://www.youtube.com/watch?v=Wz1F9FQzb_c)

## Quick Start
To build Unitd from source code use go get command and copy [unitd.conf](https://github.com/unit-io/unitd/tree/master/unitd.conf) to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd && unitd

## Usage
The unitd supports publish/subscribe to topic. You need to register a client id to connect to the unitd daemon service and generate keys for topic in order to publish or subscribe to topics. See [usage guide](https://github.com/unit-io/unitd/tree/master/docs/usage.md). 

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/unitd/blob/master/LICENSE).
