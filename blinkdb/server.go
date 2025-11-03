package blinkdb

import (
	"fmt"
	"log/slog"
	"net"
)

const defaultListenAddr = ":5001"

// Message represents an internal communication structure for commands
type Message struct {
	cmd  Command
	peer *Peer
}

// Server represents the main BlinkDB server
type Server struct {
	Config
	ln        net.Listener
	peers     map[*Peer]bool
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message
	kv        *KV
}

// Config holds the server configuration
type Config struct {
	ListenAddr string
	Username   string
	Password   string
}

// NewServer creates a new Server instance with the given configuration
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

// Start begins listening for connections and handling requests
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.loop()
	slog.Info("BlinkDB Server Running", "listenAddr", s.ListenAddr)
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
	if _, isAuth := msg.cmd.(AuthCommand); !isAuth {
		if !msg.peer.authenticated {
			msg.peer.conn.Write([]byte("-NOAUTH Authentication required - AUTH {USERNAME} {PASSWORD}\r\n"))
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
	case DeleteCommand:
		s.kv.Delete(v.key)
	case AuthCommand:
		validAuth := false
		if v.username != "" {
			validAuth = (v.username == s.Username && v.password == s.Password)
		} else {
			validAuth = (v.password == s.Password)
		}

		if validAuth {
			msg.peer.authenticated = true
			msg.peer.conn.Write([]byte("+USERNAME and PASSWORD are correct\r\n"))
		} else {
			msg.peer.authenticated = false
			msg.peer.conn.Write([]byte("-USERNAME or PASSWORD are incorrect\r\n"))
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

// Shutdown gracefully stops the server and closes all connections
func (s *Server) Shutdown() {
	close(s.quitCh)
	s.ln.Close()
	for p := range s.peers {
		p.conn.Close()
	}
}
