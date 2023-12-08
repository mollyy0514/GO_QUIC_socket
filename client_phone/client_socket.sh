#!/bin/sh

export GOCACHE=/data/go-build
export PATH=$PATH:/data/data/com.termux/files/usr/bin
export GOMODCACHE=/data/go/pkg/mod

cd /data/data/com.termux/files/home/GO_QUIC_socket
chmod +x ./socket/client_socket_phone.go 2> output.txt
# println("Hello World!")
go run ./socket/client_socket_phone.go > output.txt 2>&1 
# ./socket/client_socket_phone