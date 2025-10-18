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

type Message struct {
	cmd  Command
	peer *Peer
}

type Server struct {
	Config
	ln        net.Listener
	peers     map[*Peer]bool
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message
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
		msgCh:     make(chan Message),
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
func (s *Server) handleMessage(msg Message) error {

	switch v := msg.cmd.(type) {
	case SetCommand:
		s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

	}

	return nil
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
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
	go peer.readLoop()

}

func main() {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)

	c, err := client.New("localhost:5001")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		if err := c.Set(context.TODO(), fmt.Sprintf("foo %d", i), fmt.Sprintf("bar %d", i)); err != nil {
			log.Fatal(err)
		}
		// time.Sleep(100 * time.Millisecond)

		val, err := c.Get(context.TODO(), fmt.Sprintf("foo %d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("got this back=>", val)
	}
	time.Sleep(1 * time.Second)

}
