// Open socket for every client phone
// Since we might implement both UL&DL in the future (we only use UL for now), 
// I still assign 2 ports for each device, ports[0] for UL, ports[1] for DL

package client_phone

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	// "github.com/mollyy0514/GO_QUIC_socket/client"
	"github.com/quic-go/quic-go"
)

// const SERVER = "127.0.0.1"
// const SERVER = "192.168.1.78"
const SERVER = "140.112.20.183"

// const bufferMaxSize = 1048576          // 1mb
const PACKET_LEN = 250

func client_phone() {
	// Define command-line flags
	_host := flag.String("H", "140.112.20.183", "server ip address")
	_devices := flag.String("d", "sm00", "list of devices (space-separated)")
	_ports := flag.String("p", "", "ports to bind (space-separated)")
	_bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	_length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	_duration := flag.Int("t", 3600, "time in seconds to transmit for (default 1 hour = 3600 secs)")

	// Parse command-line arguments
	flag.Parse()

	// device := *_devices
	portsList := strings.Split(*_ports, ",")

	// set TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	// capture packets in client side
	subProcess := start_tcpdump(password)

	// connect to server IP. Session is like the socket of TCP/IP
	serverAddr := fmt.Sprintf("%s:%d", *_host, portsList[0])
	session, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

	// create a stream
	// context.Background() is similar to a channel, giving QUIC a way to communicate
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Duration to run the sending process
	duration := 30 * time.Second
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
		message := client.create_packet(uint32(euler), uint32(pi), datetimedec, microsec, uint32(seq))
		client.transmit(stream, message)
		time.Sleep(500 * time.Millisecond)
		seq++
	}
	print("times up")
	close_tcpdump(subProcess)
}


