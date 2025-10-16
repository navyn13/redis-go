package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/navyn13/redis-go/client"
)

const defaultListenAddr = ":5001"

type Server struct {
	Config
	ln        net.Listener
	peers     map[*Peer]bool
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte
	kv        *KV
}

type Config struct {
	ListenAddr string
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
		kv:        NewKV(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.loop()
	slog.Info("Server Running ", "listenAddr", s.ListenAddr)
	return s.acceptLoop()
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}
func (s *Server) handleRawMessage(rawMsg []byte) error {

	cmd, err := parseCommand(string(rawMsg))
	if err != nil {
		return err
	}

	switch v := cmd.(type) {
	case SetCommand:
		s.kv.Set(v.key, v.val)
	}

	return nil
}

func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil {
				slog.Error("raw Message Error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())
	go peer.readLoop()

}

func main() {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		client := client.New("localhost:5001")
		if err := client.Set(context.TODO(), fmt.Sprintf("foo %d", i), fmt.Sprintf("bar %d", i)); err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(2 * time.Second)
	fmt.Println(server.kv.data)

}
