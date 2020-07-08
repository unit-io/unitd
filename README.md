# unitd [![GoDoc](https://godoc.org/github.com/unit-io/unitd?status.svg)](https://pkg.go.dev/github.com/unit-io/unitd)

## Unitd is an open source messaging system for microservice, IoT and real-time internet connected devices. Unitd messaging API is built for speed and security.

The unitd is blazing fast and secure messaging system for microservices, IoT, and real-time internet connected devices. Unitd satisfy the requirements for low latency and binary messaging, it is perfect messaging system for internet of things and internet connected devices.

# About Unitd

## Key characteristics
- 100% Go
- Optimized for fast publish-subscribe
- Supports message encryption
- Supports time-to-live on message
- Supports subscribing to wildcard topics

## Unitd Clients
- [unitd-go](https://github.com/unit-io/unitd-go) Lightweight and high performance publish-subscribe messaging system - Go client library.
- [unitd-js](https://github.com/unit-io/unitd-js) High performance publish-subscribe messaging system - javascript client application.

## Tutorials and Videos
The following screen cast video demonstrate the use of [unitd-js](https://github.com/unit-io/unitd-js) javascript client to publish-subscribe messages over websocket using unitd messaging system.

[![Unitd Pubsub](https://img.youtube.com/vi/Wz1F9FQzb_c/0.jpg)](https://www.youtube.com/watch?v=Wz1F9FQzb_c)

## Quick Start
To build Unitd from source code use go get command and copy [unitd.conf](https://github.com/unit-io/unitd/tree/master/unitd.conf) to the path unitd binary is placed.

> go get -u github.com/unit-io/unitd

## Usage
The unitd supports publish-subscribe to topic. First register a client id to connect to the unitd messaging system and generate keys for topic in order to publish or subscribe to topics. See [usage guide](https://github.com/unit-io/unitd/tree/master/docs/usage.md). 

Samples are available in the examples directory for reference.

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Ltd. This project is licensed under [MIT License](https://github.com/unit-io/unitd/blob/master/LICENSE).
