package main

import (
	"context"
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

	switch c := cmd.(type) {
	case SetCommand:
		slog.Info("SET Command Received", "key", c.key, "val", c.val)
	default:
		slog.Warn("Unknown Command", "cmd", cmd)
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
	go func() {
		server := NewServer(Config{})
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)
	client := client.New("localhost:5001")

	if err := client.Set(context.TODO(), "foo", "bar"); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)

}
