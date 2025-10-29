package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"
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
	Username   string
	Password   string
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
	// Allow AUTH command without authentication
	if _, isAuth := msg.cmd.(AuthCommand); !isAuth {
		if !msg.peer.authenticated {
			msg.peer.conn.Write([]byte("-NOAUTH Authentication required\r\n"))
			return nil
		}
	}

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
	case AuthCommand:
		// Verify credentials
		validAuth := false
		if v.username != "" {
			validAuth = (v.username == s.Username && v.password == s.Password)
		} else {
			validAuth = (v.password == s.Password)
		}

		if validAuth {
			msg.peer.authenticated = true
			msg.peer.conn.Write([]byte("+OK\r\n"))
		} else {
			msg.peer.authenticated = false
			msg.peer.conn.Write([]byte("-ERR invalid username-password pair\r\n"))
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
	server := NewServer(Config{
		Username: "admin",
		Password: "admin123",
	})
	go func() {
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)

	time.Sleep(1000 * time.Second)

}
