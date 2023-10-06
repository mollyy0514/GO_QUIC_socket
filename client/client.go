package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

const serverAddr = "140.112.20.183:1234" // Change to the server's IP address

func main() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	session, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		message := []byte("Hello, server!")
		_, err := stream.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Sent: %s\n", message)
	}
}
