#!/bin/sh

export GOCACHE=/data/go-build
export PATH=$PATH:/data/data/com.termux/files/usr/bin
export GOMODCACHE=/data/go/pkg/mod

cd /data/data/com.termux/files/home/GO_QUIC_socket
chmod +x ./socket/client_socket_phone.go

go run ./socket/client_socket_phone.go -d $1 -p $2 -t $3 -b $4 -l $5
# ./socket/client_socket_phone