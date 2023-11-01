package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	// "text/template"
	"time"

	// "errors"
	"flag"
	"fmt"
	"log"
	"math/big"

	// "os"
	"os/exec"
	// "strconv"

	// "time"

	"github.com/quic-go/quic-go"
)

const bufferMaxSize = 1048576 // 1mb

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	fmt.Println("Starting server...")

	host := flag.String("host", "0.0.0.0", "Host to bind")
	quicPort := flag.Int("quic", 4242, "QUIC port to listen")

	flag.Parse()

	start_tcpdump()
	go echoQuicServer(*host, *quicPort)

	select {}
}

func handleQuicStream(stream quic.Stream) {

	idx := 0
	for {
		buf := make([]byte, bufferMaxSize)
		size, err := stream.Read(buf)
		if err != nil {
			return
		}
		fmt.Printf("Received: %s\n", buf[:size])

		// responseString := "from server" + strconv.Itoa(idx)
		// responseMsg := []byte(responseString)
		// _, err = stream.Write(responseMsg)
		// if err != nil {
		// 	panic(err)
		// }
		idx += 1
	}
}

func handleQuicSession(sess quic.Connection) {
	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			return // Using panic here will terminate the program if a new connection has not come in in a while, such as transmitting large file.
		}
		go handleQuicStream(stream)
	}
}

// Start a server that echos all data on top of QUIC
func echoQuicServer(host string, quicPort int) error {
	listener, err := quic.ListenAddr(fmt.Sprintf("%s:%d", host, quicPort), generateTLSConfig(), &quic.Config{
		KeepAlivePeriod: time.Minute * 5,
		EnableDatagrams: true,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Started QUIC server! %s:%d\n", host, quicPort)

	for {
		sess, err := listener.Accept(context.Background())
		fmt.Printf("Accepted Connection! %s\n", sess.RemoteAddr())

		if err != nil {
			return err
		}

		go handleQuicSession(sess)
	}
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if (err != nil) {
		panic(err)
	}
	template := x509.Certificate{SerialNumber:big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	kl, _ := os.OpenFile("tls_key.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

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

// func pad(bb []byte, size int) []byte {
// 	l := len(bb)
// 	if l == size {
// 		return bb
// 	}
// 	tmp := make([]byte, size)
// 	copy(tmp[size-l:], bb)
// 	return tmp
// }

func start_tcpdump() {
	cmd := exec.Command("sh", "-c", "sudo tcpdump port 4242 -w capturequic.pcap")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}