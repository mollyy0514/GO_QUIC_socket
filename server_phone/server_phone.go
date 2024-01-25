package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"GO_QUIC_socket/devices"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

const PACKET_LEN = 250
const SERVER = "0.0.0.0"
const SLEEPTIME = 2

func main() {

	// // Retrieve command-line arguments
	// args := os.Args
	// // Access the argument at index 1 (index 0 is the program name)
	// password := args[1]

	fmt.Println("Starting server...")

	// Define command-line flags
	_password := flag.String("p", "", "password")
	_devices := flag.String("d", "sm00", "list of devices (space-separated)")
	// _bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	// _length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	_duration := flag.Int("t", 300, "time in seconds to transmit for (default 1 hour = 3600 secs)")
	flag.Parse()
	// set the password for sudo

	_devices_string := *_devices
	devicesList := Get_devices(_devices_string)
	portsList := Get_Port(devicesList)
	// deviceToPort := make(map[string][]int)
	// for i, device := range devicesList {
	// 	deviceToPort[device] = []int{portsList[i][0], portsList[i][1]}
	// }
	print("deviceCnt: ", len(portsList), "\n")

	duration := *_duration

	for i := 0; i < len(portsList); i++ {
		Start_tcpdump(*_password, portsList[i][0])
		Start_tcpdump(*_password, portsList[i][1])
	}
	// Sync between goroutines.
	var wg sync.WaitGroup
	for i := 0; i < len(portsList); i++ {
		wg.Add(2)
		defer wg.Done()

		go EchoQuicServer(SERVER, portsList[i][0], true, duration)
		go EchoQuicServer(SERVER, portsList[i][1], false, duration)
	}
	wg.Wait()
}

func HandleQuicStream_ul(stream quic.Stream, quicPort int, duration int) {
	// Open or create a file to store the floats in JSON format
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	h := currentTime.Hour()
	n := currentTime.Minute()
	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
	filepath := fmt.Sprintf("./data/time_%s_%02d%02d_%d.json", date, h, n, quicPort)
	timeFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer timeFile.Close()

	for {
		buf := make([]byte, PACKET_LEN)
		ts, err := Receive(stream, buf)
		if err != nil {
			return
		}
		fmt.Printf("Received %d: %f\n", quicPort, ts)

		// Marshal the float to JSON and append it to the file
		encoder := json.NewEncoder(timeFile)
		if err := encoder.Encode(ts); err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}
	}
}

func HandleQuicStream_dl(stream quic.Stream, quicPort int, duration int) {
	seq := 1
	start_time := time.Now()
	euler := 271828
	pi := 31415926
	next_transmission_time := start_time.UnixMilli()
	for time.Since(start_time) <= time.Second * time.Duration(duration) {
		for time.Now().UnixMilli() < next_transmission_time {
			// t = time.Now().UnixNano()
		}
		t := time.Now().UnixNano()
		next_transmission_time += SLEEPTIME
		fmt.Println("server sent:", t)
		datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
		microsec := uint32(t % 1e9)    // Extract remaining microseconds

		// var message []byte
		message := Create_packet(uint32(euler), uint32(pi), datetimedec, microsec, uint32(seq))
		Transmit(stream, message)
		// time.Sleep(2 * time.Millisecond)
		seq++
	}
	message := Create_packet(uint32(euler), uint32(pi), 115, 115, uint32(seq))
	Transmit(stream, message)
}

func HandleQuicSession(sess quic.Connection, quicPort int, ul bool, duration int) {
	for {
		// create a stream to receive message, and also create a channel for communication
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			fmt.Println(err)
			return // Using panic here will terminate the program if a new connection has not come in in a while, such as transmitting large file.
		}

		if ul {
			go HandleQuicStream_ul(stream, quicPort, duration)
		} else {
			go HandleQuicStream_dl(stream, quicPort, duration)
		}
	}
}

// Start a server that echos all data on top of QUIC
func EchoQuicServer(host string, quicPort int, ul bool, duration int) error {
	quicConfig := quic.Config{
		KeepAlivePeriod: time.Minute * 5,
		EnableDatagrams: true,
		Allow0RTT:       true,
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
			filename := fmt.Sprintf("./log_%s_%02d%02d_%d_%s.qlog", date, h, n, quicPort, role)
			f, err := os.Create(filename)
			if err != nil {
				fmt.Println("cannot generate qlog file")
			}
			// handle the error
			return qlog.NewConnectionTracer(f, p, connID)
		},
	}
	listener, err := quic.ListenAddr(fmt.Sprintf("%s:%d", host, quicPort), generateTLSConfig(quicPort), &quicConfig)
	if err != nil {
		return err
	}

	fmt.Printf("Started QUIC server! %s:%d\n", host, quicPort)

	for {
		// create a session
		sess, err := listener.Accept(context.Background())
		fmt.Printf("Accepted Connection! %s\n", sess.RemoteAddr())

		if err != nil {
			return err
		}

		go HandleQuicSession(sess, quicPort, ul, duration)
	}
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig(quicPort int) *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	h := currentTime.Hour()
	n := currentTime.Minute()
	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
	keyFileName := fmt.Sprintf("./data/tls_key_%s_%02d%02d_%02d.log", date, h, n, quicPort)
	kl, _ := os.OpenFile(keyFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3"},
		KeyLogWriter: kl,
	}
}

func Start_tcpdump(password string, port int) {
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	h := currentTime.Hour()
	n := currentTime.Minute()
	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
	filepath := fmt.Sprintf("./data/capturequic_s_%s_%02d%02d_%d.pcap", date, h, n, port)
	command := fmt.Sprintf("echo %s | sudo -S tcpdump port %d -w %s", password, port, filepath)
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tcpdump start for: %02d \n", port)
}

func Get_devices(_devices_string string) []string {
	var devicesList []string
	// var serialsList []string
	if strings.Contains(_devices_string, "-") {
		pmodel := _devices_string[:2]
		start, _ := strconv.Atoi(_devices_string[2:4])
		stop, _ := strconv.Atoi(_devices_string[5:7])
		for i := start; i <= stop; i++ {
			_dev := fmt.Sprintf("%s%02d", pmodel, i)
			devicesList = append(devicesList, _dev)
			// serial := devices.Device_to_serial[_dev]
			// serialsList = append(serialsList, serial)
		}
	} else {
		devicesList = strings.Split(_devices_string, " ")
		// for _, dev := range devicesList {
		// 	serial := devices.Device_to_serial[dev]
		// 	serialsList = append(serialsList, serial)
		// }
	}

	return devicesList
}

func Get_Port(devicesList []string) [][2]int {
	var portsList [][2]int
	for _, device := range devicesList {
		// default uplink port and downlink port for each device
		ports := []int{devices.Device_to_port[device][0], devices.Device_to_port[device][1]}
		portsList = append(portsList, [2]int(ports))
	}
	return portsList
}

func Receive(stream quic.Stream, buf []byte) (float64, error) {
	_, err := stream.Read(buf)
	tsSeconds := binary.BigEndian.Uint32(buf[8:12])
	tsMicroseconds := binary.BigEndian.Uint32(buf[12:16])
	ts := float64(tsSeconds) + float64(tsMicroseconds)/1e9
	if err != nil {
		return -115, err
		// fmt.Println(err)
	}

	return ts, err
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
