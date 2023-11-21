package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	// "strings"
	"time"

	// "strings"
	// "strconv"
	// "GO_QUIC_socket/rtt_stats"
	"github.com/quic-go/quic-go"
)

// const SERVER = "127.0.0.1"
const SERVER = "192.168.1.78"
const PORT = 4242

var serverAddr string = fmt.Sprintf("%s:%d", SERVER, PORT)

// const bufferMaxSize = 1048576          // 1mb
const PACKET_LEN = 250

func main() {
	// set the password for sudo
	// Retrieve command-line arguments
	args := os.Args
	// Access the argument at index 1 (index 0 is the program name)
	password := args[1]

	// set TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	// capture packets in client side
	flag := false
	subProcess, flag := start_tcpdump(password)

	// connect to server IP. Session is like the socket of TCP/IP
	session, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil)
	if err != nil {
		fmt.Println(err)
		// log.Fatal(err)
	}
	defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

	// create a stream
	// context.Background() is similar to a channel, giving QUIC a way to communicate
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// ticker := time.NewTicker(500 * time.Millisecond)
	// defer ticker.Stop()

	// Duration to run the sending process
	duration := 30 * time.Second
	idx := 0
	start_time := time.Now()
	euler := 271828
	pi := 31415926
	for time.Since(start_time) <= time.Duration(duration) && flag {
		// for time.Since(start_time) < 600*time.Millisecond {
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

		// add random additional data to 250 bytes
		msgLength := len(message)
		if msgLength < PACKET_LEN {
			randomBytes := make([]byte, PACKET_LEN-msgLength)
			rand.Read(randomBytes)
			message = append(message, randomBytes...)
		}

		_, err = stream.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		idx++
	}
	print("times up")
	close_tcpdump(subProcess)

	// responseBuf := make([]byte, bufferMaxSize)
	// size, err := stream.Read(responseBuf)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("Received: %s\n", responseBuf[:size])
	// }
}

func start_tcpdump(password string) (*exec.Cmd, bool) {
	// flag := false
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	filepath := fmt.Sprintf("./data/capturequic_c_%d%d%d.pcap", y, m, d)
	command := fmt.Sprintf("echo %s | sudo -S tcpdump port %d -w %s", password, PORT, filepath)
	subProcess := exec.Command("sh", "-c", command)

	// subProcess.Stdin = strings.NewReader(password)

	err := subProcess.Start()
	if err != nil {
		log.Fatal(err)
	}

	return subProcess, true
}

func close_tcpdump(cmd *exec.Cmd) {
	quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
}
