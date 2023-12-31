package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
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
)

const packet_length = 250
const SERVER = "0.0.0.0"

func main() {

	// // Retrieve command-line arguments
	// args := os.Args
	// // Access the argument at index 1 (index 0 is the program name)
	// password := args[1]

	fmt.Println("Starting server...")

	// Define command-line flags
	_password := flag.String("p", "", "password")
	_devices := flag.String("d", "sm00-01", "list of devices (space-separated)")
	// _bitrate := flag.String("b", "1M", "target bitrate in bits/sec (0 for unlimited)")
	// _length := flag.String("l", "250", "length of buffer to read or write in bytes (packet size)")
	// _duration := flag.Int("t", 3600, "time in seconds to transmit for (default 1 hour = 3600 secs)")
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

	for i := 0; i < len(portsList); i++ {
		Start_tcpdump(*_password, devicesList[i], portsList[i][0])
	}
	// Sync between goroutines.
	var wg sync.WaitGroup
	for i := 0; i < len(portsList); i++ {
		wg.Add(1)
		defer wg.Done()
		// go func() {
		go EchoQuicServer(SERVER, portsList[i][0])
		// }()
		// select {}
	}
	wg.Wait()
}

func HandleQuicStream(stream quic.Stream, quicPort int) {

	seq := 0
	for {
		buf := make([]byte, packet_length)
		ts, err := Receive(stream, buf)
		if err != nil {
			return
		}
		fmt.Printf("Received %d: %f\n", quicPort, ts)

		// sending response to client
		// responseString := "server received!"
		// responseMsg := []byte(responseString)
		// response(stream, responseMsg)

		seq += 1
	}
}

func HandleQuicSession(sess quic.Connection, quicPort int) {
	for {
		// create a stream to receive message, and also create a channel for communication
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			return // Using panic here will terminate the program if a new connection has not come in in a while, such as transmitting large file.
		}
		go HandleQuicStream(stream, quicPort)
	}
}

// Start a server that echos all data on top of QUIC
func EchoQuicServer(host string, quicPort int) error {
	listener, err := quic.ListenAddr(fmt.Sprintf("%s:%d", host, quicPort), generateTLSConfig(quicPort), &quic.Config{
		KeepAlivePeriod: time.Minute * 5,
		EnableDatagrams: true,
	})
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

		go HandleQuicSession(sess, quicPort)
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

	keyFileName := fmt.Sprintf("./data/tls_key_%02d.log", quicPort)
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

func Start_tcpdump(password string, device string, port int) {
	currentTime := time.Now()
	y := currentTime.Year()
	m := currentTime.Month()
	d := currentTime.Day()
	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
	filepath := fmt.Sprintf("./data/capturequic_s_%s_%s.pcap", date, device)
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
	ts := float64(tsSeconds) + float64(tsMicroseconds)/1e10
	if err != nil {
		return -115, err
		// fmt.Println(err)
	}

	return ts, err
}
