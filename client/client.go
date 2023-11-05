package main

import (
	"context"
	"os/exec"
	// "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"

	// "strconv"
	"time"

	"github.com/quic-go/quic-go"
)

const serverAddr = "192.168.1.78:4242" // Change to the server's IP address
// const bufferMaxSize = 1048576          // 1mb
const PACKET_LEN = 250
const SERVER_TIME_SYNC = "syncing from server"

func main() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	// calture packets in client side
	start_tcpdump()
	
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

	// Duration to run the sending process
	// duration := 1 * time.Minute

	idx := 0
	start_time := time.Now()
	euler := 271828
	pi := 31415926

	// for time.Since(start_time) <= time.Duration(duration) {
	for time.Since(start_time) < 600*time.Millisecond {
		// str := "Hello, server "+ time.Since(start_time).String()
		// message := []byte(str)
		t := time.Now().UnixNano() // Time in milliseconds
		print(t, "\n")
		datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
		microsec := uint32(t % 1e9)    // Extract remaining microseconds

		var message []byte
		message = append(message, make([]byte, 4)...)
		binary.BigEndian.PutUint32(message[:4], uint32(euler))
		message = append(message, make([]byte, 4)...)
		binary.BigEndian.PutUint32(message[4:8], uint32(pi))
		message = append(message, make([]byte, 4)...)
		binary.BigEndian.PutUint32(message[8:12], datetimedec)
		message = append(message, make([]byte, 4)...)
		binary.BigEndian.PutUint32(message[12:16], microsec)

		// add random additinal data to 250 bytes
		// msgLength := len(message)
		// if msgLength < PACKET_LEN {
		// 	randomBytes := make([]byte, PACKET_LEN-msgLength)
		// 	rand.Read(randomBytes)
		// 	message = append(message, randomBytes...)
		// }

		_, err = stream.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		idx++
	}
	// print("times up")

	// responseBuf := make([]byte, bufferMaxSize)
	// size, err := stream.Read(responseBuf)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("Received: %s\n", responseBuf[:size])
	// }
}

func start_tcpdump() {
	cmd := exec.Command("sh", "-c", "sudo tcpdump port 4242 -w ./data/capturequic_c.pcap")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
