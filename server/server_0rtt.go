package main

// import (
// 	"context"
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"encoding/binary"
// 	"encoding/pem"
// 	"os"
// 	"time"

// 	// "errors"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"math/big"
// 	"os/exec"

// 	// "strconv"
// 	"github.com/quic-go/quic-go"
// 	"github.com/quic-go/quic-go/logging"
// 	"github.com/quic-go/quic-go/qlog"
// )

// // const bufferMaxSize = 1048576 // 1mb
// const packet_length = 250
// const SERVER = "0.0.0.0"
// const PORT = 4242

// // We start a server echoing data on the first stream the client opens,
// // then connect with a client, send the message, and wait for its receipt.
// func main() {
// 	fmt.Println("Starting server...")

// 	host := flag.String("host", SERVER, "Host to bind")
// 	quicPort := flag.Int("quic", PORT, "QUIC port to listen")
// 	flag.Parse()

// 	// Start_server_tcpdump()

// 	go EchoQuicServer(*host, *quicPort)
// 	select {}

// 	// go EchoServer(*host, *quicPort)
// 	// select {}
// }

// func EchoServer(host string, quicPort int) error {
// 	quicConfig := quic.Config{
// 		KeepAlivePeriod: time.Minute * 5,
// 		EnableDatagrams: true,
// 		Allow0RTT:       true,
// 		// Tracer: func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
// 		// 	role := "server"
// 		// 	if p == logging.PerspectiveClient {
// 		// 		role = "client"
// 		// 	}
// 		// 	filename := fmt.Sprintf("./log_%s_%s.qlog", connID, role)
// 		// 	f, err := os.Create(filename)
// 		// 	if err != nil {
// 		// 		fmt.Println("cannot generate qlog file")
// 		// 	}
// 		// 	// handle the error
// 		// 	return qlog.NewConnectionTracer(f, p, connID)
// 		// },
// 	}
// 	// ListenAddrEarly supports 0rtt
// 	listener, err := quic.ListenAddrEarly(fmt.Sprintf("%s:%d", host, quicPort), GenerateTLSConfig(), &quicConfig)
// 	if err != nil {
// 		return err
// 	}
// 	defer listener.Close()

// 	fmt.Printf("Started QUIC server! %s:%d\n", host, quicPort)

// 	for {
// 		conn, err := listener.Accept(context.Background())
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}

// 		stream, err := conn.AcceptStream(context.Background())
// 		if err != nil {
// 			// panic(err)
// 			fmt.Println(err)
// 			return err
// 		}
// 		defer stream.Close()

// 		state := conn.ConnectionState()
// 		log.Printf("Accepted QUIC connection from %v, 0rtt = %v, handshake = %v, TLSresume = %v, proto = %s\n",
// 			conn.RemoteAddr(),
// 			state.Used0RTT,
// 			state.TLS.HandshakeComplete,
// 			state.TLS.DidResume,
// 			state.TLS.NegotiatedProtocol)

// 		for {
// 			buf := make([]byte, packet_length)
// 			ts, err := Server_receive(stream, buf)
// 			if err != nil {
// 				return err
// 			}
// 			fmt.Printf("Received: %f\n", ts)
// 		}
// 	}

// }

// func HandleQuicStream(stream quic.Stream) {

// 	seq := 0
// 	for {
// 		buf := make([]byte, packet_length)
// 		ts, err := Server_receive(stream, buf)
// 		if err != nil {
// 			return
// 		}
// 		fmt.Printf("Received: %f\n", ts)

// 		seq += 1
// 	}
// 	// written, err := io.Copy(loggingWriter{stream}, stream)
// 	// if err != nil {
// 	// 	println(err)
// 	// }
// 	// fmt.Printf("written %d", written)
// }

// func HandleQuicSession(sess quic.Connection) {
// 	for {
// 		// create a stream to receive message, and also create a channel for communication
// 		stream, err := sess.AcceptStream(context.Background())
// 		if err != nil {
// 			fmt.Println(err)
// 			return // Using panic here will terminate the program if a new connection has not come in in a while, such as transmitting large file.
// 		}
// 		// print remote address & 0rtt
// 		state := sess.ConnectionState()
// 		log.Printf("Accepted QUIC connection from %v, 0rtt = %v, handshake = %v, TLSresume = %v\n",
// 			sess.RemoteAddr(),
// 			state.Used0RTT,
// 			state.TLS.HandshakeComplete,
// 			state.TLS.DidResume)

// 		go HandleQuicStream(stream)
// 	}
// }

// // Start a server that echos all data on top of QUIC
// func EchoQuicServer(host string, quicPort int) error {
// 	quicConfig := quic.Config{
// 		KeepAlivePeriod: time.Minute * 5,
// 		EnableDatagrams: true,
// 		Allow0RTT:       true,
		
// 		Tracer: func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
// 			role := "server"
// 			if p == logging.PerspectiveClient {
// 				role = "client"
// 			}
// 			filename := fmt.Sprintf("./log_%s_%s.qlog", connID, role)
// 			f, err := os.Create(filename)
// 			if err != nil {
// 				fmt.Println("cannot generate qlog file")
// 			}
// 			// handle the error
// 			return qlog.NewConnectionTracer(f, p, connID)
// 		},
// 	}
// 	// ListenAddrEarly supports 0rtt
// 	listener, err := quic.ListenAddrEarly(fmt.Sprintf("%s:%d", host, quicPort), GenerateTLSConfig(), &quicConfig)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("Started QUIC server! %s:%d\n", host, quicPort)

// 	for {
// 		// create a session
// 		sess, err := listener.Accept(context.Background())
// 		fmt.Printf("Accepted Connection! %s\n", sess.RemoteAddr())

// 		if err != nil {
// 			return err
// 		}

// 		go HandleQuicSession(sess)
// 	}
// }

// // Setup a bare-bones TLS config for the server
// func GenerateTLSConfig() *tls.Config {
// 	key, err := rsa.GenerateKey(rand.Reader, 1024)
// 	if err != nil {
// 		panic(err)
// 	}
// 	template := x509.Certificate{SerialNumber: big.NewInt(1)}
// 	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
// 	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
// 	kl, _ := os.OpenFile("tls_key.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

// 	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return &tls.Config{
// 		Certificates: []tls.Certificate{tlsCert},
// 		NextProtos:   []string{"h3"},
// 		KeyLogWriter: kl,
// 	}
// }

// func Start_server_tcpdump() {
// 	currentTime := time.Now()
// 	y := currentTime.Year()
// 	m := currentTime.Month()
// 	d := currentTime.Day()
// 	date := fmt.Sprintf("%02d%02d%02d", y, m, d)
// 	filepath := fmt.Sprintf("./data/capturequic_s_%s.pcap", date)
// 	command := fmt.Sprintf("sudo tcpdump port %d -w %s", PORT, filepath)
// 	cmd := exec.Command("sh", "-c", command)
// 	err := cmd.Start()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func Server_receive(stream quic.Stream, buf []byte) (float64, error) {
// 	_, err := stream.Read(buf)
// 	tsSeconds := binary.BigEndian.Uint32(buf[8:12])
// 	tsMicroseconds := binary.BigEndian.Uint32(buf[12:16])
// 	ts := float64(tsSeconds) + float64(tsMicroseconds)/1e10
// 	if err != nil {
// 		return -115, err
// 		// fmt.Println(err)
// 	}

// 	return ts, err
// }
