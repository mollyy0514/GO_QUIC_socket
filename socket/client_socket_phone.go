// Open socket for every client phone
// Since we might implement both UL&DL in the future (we only use UL for now),
// I still assign 2 ports for each device, ports[0] for UL, ports[1] for DL

package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	// "os"
	"os/exec"
	// "os/signal"
	"strings"
	// "syscall"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

// const SERVER = "127.0.0.1"
// const SERVER = "192.168.1.78"
const SERVER = "140.112.20.183"

// const bufferMaxSize = 1048576          // 1mb
const PACKET_LEN = 250
const SLEEPTIME = 2

// func Socket(_host *string, _devices *string, _ports *string, _bitrate *string, _length *string, _duration *int) {
func main() {
	// Define command-line flags
	_host := flag.String("H", "140.112.20.183", "server ip address")
	_devices := flag.String("d", "sm00", "list of devices (space-separated)")
	_ports := flag.String("p", "4200,4201", "ports to bind (space-separated)")
	_bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	_length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	_duration := flag.Int("t", 300, "time in seconds to transmit for (default 1 hour = 3600 secs)")
	// Parse command-line arguments
	flag.Parse()
	fmt.Printf("info %s %s %s %s %s %d \n", *_host, *_devices, *_ports, *_bitrate, *_length, *_duration)

	// ports := *_ports
	portsList := strings.Split(*_ports, ",")
	
	duration := *_duration

	var PORT_UL int
	var PORT_DL int
	if (len(portsList) == 2) {
		PORT_UL, _ = strconv.Atoi(portsList[0])
		PORT_DL, _ = strconv.Atoi(portsList[1])
	} else {
		fmt.Println("port missing!")
	}
	var serverAddr_ul string = fmt.Sprintf("%s:%d", SERVER, PORT_UL)
	var serverAddr_dl string = fmt.Sprintf("%s:%d", SERVER, PORT_DL)

	/* ---------- TCPDUMP ---------- */
	// subp1 := Start_client_tcpdump(portsList[0])
	// subp2 := Start_client_tcpdump(portsList[1])
	// time.Sleep(1 * time.Second) // sleep 1 sec to ensure the whle handshake process is captured
	/* ---------- TCPDUMP ---------- */

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) { // capture packets in client side
			if i == 0 {		// UPLINK
				// set generate configs
				tlsConfig := GenTlsConfig()
				quicConfig := GenQuicConfig(PORT_UL)

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
				defer cancel()
				// connect to server IP. Session is like the socket of TCP/IP
				session_ul, err := quic.DialAddr(ctx, serverAddr_ul, tlsConfig, &quicConfig)
				if err != nil {
					fmt.Println("err: ", err)
				}
				defer session_ul.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")
				// create a stream_ul
				// context.Background() is similar to a channel, giving QUIC a way to communicate
				stream_ul, err := session_ul.OpenStreamSync(context.Background())
				if err != nil {
					log.Fatal(err)
				}
				defer stream_ul.Close()

				Client_send(stream_ul, duration)
				time.Sleep(1 * time.Second)
				session_ul.CloseWithError(0, "ul times up")
				/* ---------- TCPDUMP ---------- */
				// Close_tcpdump(subp1)
				/* ---------- TCPDUMP ---------- */
			} else {	// DOWNLINK
				// set generate configs
				tlsConfig := GenTlsConfig()
				quicConfig := GenQuicConfig(PORT_DL)

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
				defer cancel()
				// connect to server IP. Session is like the socket of TCP/IP
				session_dl, err := quic.DialAddr(ctx, serverAddr_dl, tlsConfig, &quicConfig)
				if err != nil {
					fmt.Println("err: ", err)
				}
				defer session_dl.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")
				// create a stream_dl
				// context.Background() is similar to a channel, giving QUIC a way to communicate
				stream_dl, err := session_dl.OpenStreamSync(context.Background())
				if err != nil {
					log.Fatal(err)
				}
				defer stream_dl.Close()

				// Open or create a file to store the floats in JSON format
				currentTime := time.Now()
				y := currentTime.Year()
				m := currentTime.Month()
				d := currentTime.Day()
				h := currentTime.Hour()
				n := currentTime.Minute()
				date := fmt.Sprintf("%02d%02d%02d", y, m, d)
				filepath := fmt.Sprintf("/sdcard/pcapdir/time_%s_%02d%02d_%d.json", date, h, n, PORT_DL)
				timeFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer timeFile.Close()

				var message []byte
				t := time.Now().UnixNano() // Time in milliseconds
				fmt.Println("client create time: ", t)
				datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
				microsec := uint32(t % 1e9)    // Extract remaining microseconds
				message = append(message, make([]byte, 4)...)
				binary.BigEndian.PutUint32(message[:4], datetimedec)
				message = append(message, make([]byte, 4)...)
				binary.BigEndian.PutUint32(message[4:8], microsec)
				SendStartPacket(stream_dl, message)

				for {
					buf := make([]byte, PACKET_LEN)
					ts, err := Client_receive(stream_dl, buf)
					if (ts == -115) {
						session_dl.CloseWithError(0, "dl times up")
						/* ---------- TCPDUMP ---------- */
						// Close_tcpdump(subp2)
						/* ---------- TCPDUMP ---------- */
					}
					if err != nil {
						return
					}
					fmt.Printf("client received: %f\n", ts)

					// Marshal the float to JSON and append it to the file
					encoder := json.NewEncoder(timeFile)
					if err := encoder.Encode(ts); err != nil {
						fmt.Println("Error encoding JSON:", err)
						return
					}
				}
				
			}
		}(i)
	}
	wg.Wait()
}

func Start_client_tcpdump(port string) *exec.Cmd {
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	h := currentTime.Hour()
	n := currentTime.Minute()
	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
	filepath := fmt.Sprintf("/sdcard/pcapdir/ul_c_%s_%02d%02d_%s.pcap", date, h, n, port)
	command := fmt.Sprintf("su -c tcpdump port %s -w %s", port, filepath)
	subProcess := exec.Command("sh", "-c", command)
	err := subProcess.Start()
	fmt.Printf("file created! \n")
	if err != nil {
		log.Fatal(err)
	}

	return subProcess
}

func GenTlsConfig() *tls.Config {
	// set TLS
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}
}

func GenQuicConfig(port int) quic.Config {
	return quic.Config{
		Allow0RTT: true,
		Tracer: func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			role := "server"
			if p == logging.PerspectiveClient {
				role = "client"
			}
			currentTime := time.Now()
			y := currentTime.Year()
			m := currentTime.Month()
			d := currentTime.Day()
			h := currentTime.Hour()
			n := currentTime.Minute()
			date := fmt.Sprintf("%02d%02d%02d", y, m, d)
			filename := fmt.Sprintf("/sdcard/pcapdir/log_%s_%02d%02d_%d_%s.qlog", date, h, n, port, role)
			f, err := os.Create(filename)
			if err != nil {
				fmt.Println("cannot generate qlog file")
			}
			// handle the error
			return qlog.NewConnectionTracer(f, p, connID)
		},
	}
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

func SendStartPacket(stream quic.Stream, message []byte) {
	_, err := stream.Write(message)
	if err != nil {
		log.Fatal(err)
	}
}

func Client_send(stream quic.Stream, duration int) {
	seq := 1
	start_time := time.Now()
	euler := 271828
	pi := 31415926
	next_transmission_time := start_time.UnixMilli()
	for time.Since(start_time) <= time.Second * time.Duration(duration) {
		for time.Now().UnixMilli() < next_transmission_time {
			// t = time.Now().UnixNano()
		}
		next_transmission_time += SLEEPTIME
		t := time.Now().UnixNano() // Time in milliseconds
		fmt.Println("client sent: ", t)
		datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
		microsec := uint32(t % 1e9)    // Extract remaining microseconds

		// var message []byte
		message := Create_packet(uint32(euler), uint32(pi), datetimedec, microsec, uint32(seq))
		SendStartPacket(stream, message)
		// time.Sleep(2 * time.Millisecond)
		seq++
	}
}

func Client_receive(stream quic.Stream, buf []byte) (float64, error) {
	_, err := stream.Read(buf)
	tsSeconds := binary.BigEndian.Uint32(buf[8:12])
	tsMicroseconds := binary.BigEndian.Uint32(buf[12:16])
	var ts float64
	if (tsSeconds == 115 && tsMicroseconds == 115) {
		return -115, err
	} else {
		ts = float64(tsSeconds) + float64(tsMicroseconds)/1e9
	}

	if err != nil {
		return -1103, err
	}

	return ts, err
}
