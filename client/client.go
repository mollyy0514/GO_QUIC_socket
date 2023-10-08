package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/quic-go/quic-go"
)

const serverAddr = "192.168.1.79:4242" // Change to the server's IP address
const bufferMaxSize = 1048576 // 1mb

func main() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	session, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil)
	if err != nil {
		fmt.Println(err)
		// log.Fatal(err)
	}
	defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	idx := 0
	for range ticker.C {	
		str := "Hello, server " + strconv.Itoa(idx)
		message := []byte(str)
		_, err := stream.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Sent: %s\n", message)
		idx += 1

		responseBuf := make([]byte, bufferMaxSize)
		size, err := stream.Read(responseBuf)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Received: %s\n", responseBuf[:size])
	}
}
