// Open socket for every client phone
// Since we might implement both UL&DL in the future (we only use UL for now),
// I still assign 2 ports for each device, ports[0] for UL, ports[1] for DL

package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"

	"flag"
	"fmt"
	"log"
	// "os"
	"os/exec"
	// "os/signal"
	"strings"
	// "syscall"
	"time"

	"github.com/quic-go/quic-go"
)

// const SERVER = "127.0.0.1"
// const SERVER = "192.168.1.78"
const SERVER = "140.112.20.183"

// const bufferMaxSize = 1048576          // 1mb
const PACKET_LEN = 250

// func Socket(_host *string, _devices *string, _ports *string, _bitrate *string, _length *string, _duration *int) {
func main() {
	// Define command-line flags
	_host := flag.String("H", "192.168.1.79", "server ip address")
	_devices := flag.String("d", "sm00", "list of devices (space-separated)")
	_ports := flag.String("p", "3200", "ports to bind (space-separated)")
	_bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	_length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	_duration := flag.Int("t", 3600, "time in seconds to transmit for (default 1 hour = 3600 secs)")
	// Parse command-line arguments
	flag.Parse()
	fmt.Printf("info %s %s %s %s %s %d \n", *_host, *_devices, *_ports, *_bitrate, *_length, *_duration)

	// ports := *_ports
	portsList := strings.Split(*_ports, ",")

	// set TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	// capture packets in client side
	subProcess := Start_tcpdump(portsList)

	// connect to server IP. Session is like the socket of TCP/IP
	serverAddr := fmt.Sprintf("%s:%s", *_host, portsList[0])
	fmt.Printf("serverAddr: %s \n", serverAddr)
	session, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil)
	if err != nil {
		fmt.Print("err: ")
		fmt.Println(err)
		return
	}
	fmt.Print("dialed \n")
	defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

	// create a stream
	// context.Background() is similar to a channel, giving QUIC a way to communicate
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Duration to run the sending process
	duration := 5 * time.Second
	seq := 1
	start_time := time.Now()
	euler := 271828
	pi := 31415926
	for time.Since(start_time) <= time.Duration(duration) {

		t := time.Now().UnixNano() // Time in milliseconds
		print(t, "\n")
		datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
		microsec := uint32(t % 1e9)    // Extract remaining microseconds

		// var message []byte
		message := Create_packet(uint32(euler), uint32(pi), datetimedec, microsec, uint32(seq))
		Transmit(stream, message)
		time.Sleep(500 * time.Millisecond)
		seq++
	}
	print("times up")
	Close_tcpdump(subProcess)
}

func Start_tcpdump(portsList []string) *exec.Cmd {
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	filepath := fmt.Sprintf("./data/capturequic_c_%02d%02d%02d.pcap", y, m, d)
	// command := fmt.Sprintf("echo wmnlab | sudo -S tcpdump port %s -w %s", portsList[0], filepath)
	command := fmt.Sprintf("tcpdump port %s -w %s", portsList[0], filepath)
	subProcess := exec.Command("su", "-c", command)
	err := subProcess.Start()
	fmt.Printf("file created! \n")
	if err != nil {
		log.Fatal(err)
	}

	return subProcess
}

func Close_tcpdump(cmd *exec.Cmd) {
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	fmt.Print(cmd)
}

func Create_packet(euler uint32, pi uint32, datetimedec uint32, microsec uint32, seq uint32) []byte {
	var message []byte
	message = append(message, make([]byte, 4)...)
	binary.BigEndian.PutUint32(message[:4], euler)
	message = append(message, make([]byte, 4)...)
	binary.BigEndian.PutUint32(message[4:8], pi)
	message = append(message, make([]byte, 4)...)
	binary.BigEndian.PutUint32(message[8:12], datetimedec)
	message = append(message, make([]byte, 4)...)
	binary.BigEndian.PutUint32(message[12:16], microsec)
	message = append(message, make([]byte, 4)...)
	binary.BigEndian.PutUint32(message[16:20], seq)

	// add random additional data to 250 bytes
	msgLength := len(message)
	if msgLength < PACKET_LEN {
		randomBytes := make([]byte, PACKET_LEN-msgLength)
		rand.Read(randomBytes)
		message = append(message, randomBytes...)
	}

	return message
}

func Transmit(stream quic.Stream, message []byte) {
	_, err := stream.Write(message)
	if err != nil {
		log.Fatal(err)
	}
}
