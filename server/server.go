package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"crypto/tls"

	"github.com/quic-go/quic-go"
)

const serverAddr = "0.0.0.0:1234" // Change to the desired server IP address

func handleClient(session quic.Connection) {
	for {
		stream, err := session.AcceptStream(session.Context())
		if err != nil {
			log.Println(err)
			return
		}

		go func(s quic.Stream) {
			defer s.Close()

			buf := make([]byte, 1024)
			for {
				n, err := s.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}
				fmt.Printf("Received: %s\n", buf[:n])
			}
		}(stream)
	}
}

func main() {
	// Listen for QUIC connections
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1234})
	if err != nil {
		log.Fatal(err)
	}
	defer udpConn.Close()

	tr := quic.Transport{
		Conn: udpConn,
	}
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}
	quicConf := &quic.Config{
		// KeepAlive: true,
	}

	quicListener, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		log.Fatal(err)
	}
	defer quicListener.Close()

	fmt.Printf("Listening on %s...\n", serverAddr)

	go func() {
		fmt.Println("Listening on...")
		for {
			conn, err := quicListener.Accept(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			go handleClient(conn)
		}
	}()
}
