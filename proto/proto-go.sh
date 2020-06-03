#!/bin/bash
go generate protoc --proto_path=../proto --go_out=plugins=grpc:../proto ../proto/unitd.proto