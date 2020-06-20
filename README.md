# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://pkg.go.dev/github.com/unit-io/unitd)

## Unitd is a real-time messaging service for IoT connected devices, it supports pubsub using GRPC client or MQTT client over tcp or websocket. 

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
To run unitdb as daemon service start [unitd](https://github.com/unit-io/unitd) application and copy unitd.conf to the path unitd binary is placed. Unitd supports pubsub using GRPC client or MQTT client to connect to service using tcp or websocket.
- [unitd-go](https://github.com/unit-io/unitd-go) is Go client to pubsub messages using GRPC application
- [unitd-ws](https://github.com/unit-io/unitd-ws) is javascript client to pubsub messages using MQTT over websocket 


## Quick Start
To build Unitd from source code use go get command and copy [unitd.conf](https://github.com/unit-io/unitd/tree/master/unitd.conf) to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd && unitd

## Usage
The unitd supports publish/subscribe to topic. You need to register a client id to connect to the unitd daemon service and generate keys for topic in order to publish or subscribe to topics. See [usage guide](https://github.com/unit-io/unitd/tree/master/docs/usage.md). 

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/unitd/blob/master/LICENSE).
