package main

// import (
// 	"context"
// 	"crypto/rand"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"encoding/binary"
// 	"fmt"

// 	// "io"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/quic-go/quic-go"
// 	"github.com/quic-go/quic-go/logging"
// 	"github.com/quic-go/quic-go/qlog"
// )

// const SERVER = "127.0.0.1"

// // const SERVER = "192.168.1.79" // MacBook Pro M1 local IP
// // const SERVER = "192.168.1.78" // wmnlab local IP
// // const SERVER = "140.112.20.183"  // 249 public IP
// const PORT = 4242

// var serverAddr string = fmt.Sprintf("%s:%d", SERVER, PORT)

// const PACKET_LEN = 250

// func main() {

// 	pool, err := x509.SystemCertPool()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// println("pool", pool)

// 	// set TLS
// 	tlsConfig := &tls.Config{
// 		InsecureSkipVerify: false,
// 		RootCAs: pool,
// 		// KeyLogWriter: keyLog,
// 		NextProtos:         []string{"h3"},
// 		ClientSessionCache: tls.NewLRUClientSessionCache(100),
// 	}
// 	// var put <-chan struct{}
// 	// tlsConfig.ClientSessionCache, put = newSessionCache(tls.NewLRUClientSessionCache(1))
// 	// puts := make(chan string, 100)
// 	// cache := tlsConfig.ClientSessionCache
// 	// if cache == nil {
// 	// 	cache = tls.NewLRUClientSessionCache(100)
// 	// }
// 	// tlsConfig.ClientSessionCache = newClientSessionCache(cache, make(chan string, 100), puts)

// 	// add 0rtt to quicConfig
// 	quicConfig := quic.Config{
// 		Allow0RTT: true,
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
// 		TokenStore: quic.NewLRUTokenStore(100, 100),
// 	}

// 	for i := 0; i < 3; i++ {
// 		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
// 		defer cancel()

// 		// connect to server IP. Session is like the socket of TCP/IP
// 		session, err := quic.DialAddrEarly(ctx, serverAddr, tlsConfig, &quicConfig)
// 		if err != nil {
// 			fmt.Println("err: ", err)
// 		}
// 		// defer session.CloseWithError(quic.ApplicationErrorCode(501), "hi you have an error")

// 		// create a stream
// 		// context.Background() is similar to a channel, giving QUIC a way to communicate
// 		stream, err := session.OpenStreamSync(context.Background())
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		// defer stream.Close()

// 		// select {
// 		// 	case <-time.NewTimer(10 * time.Second).C:
// 		// 		log.Printf("expected to receive a session ticket within 10 seconds")
// 		// 	case <-put:
// 		// 		log.Println("GET session ticket ", put)
// 		// }

// 		// Duration to run the sending process
// 		duration := 5 * time.Second
// 		seq := 1
// 		start_time := time.Now()
// 		euler := 271828
// 		pi := 31415926
// 		for time.Since(start_time) <= time.Duration(duration) {

// 			t := time.Now().UnixNano() // Time in milliseconds
// 			print(t, "\n")
// 			datetimedec := uint32(t / 1e9) // Extract seconds from milliseconds
// 			microsec := uint32(t % 1e9)    // Extract remaining microseconds

// 			// var message []byte
// 			message := Create_packet(uint32(euler), uint32(pi), datetimedec, microsec, uint32(seq))
// 			Transmit(stream, message)
// 			time.Sleep(500 * time.Millisecond)
// 			seq++
// 		}
// 		print("times up \n")

// 		state := session.ConnectionState()
// 		fmt.Printf("[4242 same]i:%d, used0RTT:%v\n", i, state.Used0RTT)

// 		stream.Close()
// 		session.CloseWithError(0, "there is an error!")

// 	}

// }

// func Create_packet(euler uint32, pi uint32, datetimedec uint32, microsec uint32, seq uint32) []byte {
// 	var message []byte
// 	message = append(message, make([]byte, 4)...)
// 	binary.BigEndian.PutUint32(message[:4], euler)
// 	message = append(message, make([]byte, 4)...)
// 	binary.BigEndian.PutUint32(message[4:8], pi)
// 	message = append(message, make([]byte, 4)...)
// 	binary.BigEndian.PutUint32(message[8:12], datetimedec)
// 	message = append(message, make([]byte, 4)...)
// 	binary.BigEndian.PutUint32(message[12:16], microsec)
// 	message = append(message, make([]byte, 4)...)
// 	binary.BigEndian.PutUint32(message[16:20], seq)

// 	// add random additional data to 250 bytes
// 	msgLength := len(message)
// 	if msgLength < PACKET_LEN {
// 		randomBytes := make([]byte, PACKET_LEN-msgLength)
// 		rand.Read(randomBytes)
// 		message = append(message, randomBytes...)
// 	}

// 	return message
// }

// func Transmit(stream quic.Stream, message []byte) {
// 	_, err := stream.Write(message)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// // type sessionCache struct {
// // 	tls.ClientSessionCache
// // 	put chan<- struct{}
// // }
// // func newSessionCache(c tls.ClientSessionCache) (tls.ClientSessionCache, <-chan struct{}) {
// // 	put := make(chan struct{}, 100)
// // 	return &sessionCache{ClientSessionCache: c, put: put}, put
// // }

// // func (c *sessionCache) Put(key string, cs *tls.ClientSessionState) {
// // 	c.ClientSessionCache.Put(key, cs)
// // 	c.put <- struct{}{}
// // }

// // type clientSessionCache struct {
// // 	cache tls.ClientSessionCache

// // 	gets chan<- string
// // 	puts chan<- string
// // }

// // func newClientSessionCache(cache tls.ClientSessionCache, gets, puts chan<- string) *clientSessionCache {
// // 	return &clientSessionCache{
// // 		cache: cache,
// // 		gets:  gets,
// // 		puts:  puts,
// // 	}
// // }

// // var _ tls.ClientSessionCache = &clientSessionCache{}

// // func (c *clientSessionCache) Get(sessionKey string) (*tls.ClientSessionState, bool) {
// // 	session, ok := c.cache.Get(sessionKey)
// // 	c.gets <- sessionKey
// // 	return session, ok
// // }

// // func (c *clientSessionCache) Put(sessionKey string, cs *tls.ClientSessionState) {
// // 	c.cache.Put(sessionKey, cs)
// // 	c.puts <- sessionKey
// // }
