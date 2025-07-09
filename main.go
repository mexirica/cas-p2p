package main

import (
	"bytes"
	"fmt"
	"github.com/mexirica/cas-p2p/internal/server"
	"github.com/mexirica/cas-p2p/internal/store"
	"github.com/mexirica/cas-p2p/pkg/crypto"
	"github.com/mexirica/cas-p2p/pkg/p2p"
	"io"
	"log"
	"strings"
	"time"
)

var EncKey = crypto.DeriveKey("super-secret-passphrase")

func makeServer(listenAddr string, nodes ...string) *server.FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := server.FileServerOpts{
		EncKey:            EncKey,
		StorageRoot:       strings.TrimLeft(listenAddr, ":") + "_network",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := server.NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	// Create three servers, two of them are peers and one is a client
	// that will store and retrieve files from the other two.
	s1 := makeServer(":8080", "")
	s2 := makeServer(":8081", "")
	s3 := makeServer(":8082", ":8080", ":8081")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	key := "test_file.txt"
	data := bytes.NewReader([]byte("Think that here is a big file, with a thousand of lines"))
	s1.Store(key, data)

	r, err := s2.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("content received: %s", string(b))
}
