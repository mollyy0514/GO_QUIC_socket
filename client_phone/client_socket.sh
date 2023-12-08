#!/bin/bash

export GOCACHE=/data/go-build
export PATH=$PATH:/data/data/com.termux/files/usr/bin
export GOMODCACHE=/data/go/pkg/mod

cd /data/data/com.termux/files/home/GO_QUIC_socket
chmod +x ./socket/client_socket_phone.go
# println("Hello World!")
go run ./socket/client_socket_phone.go
# ./socket/client_socket_phone