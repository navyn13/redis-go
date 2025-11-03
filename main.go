package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const defaultListenAddr = ":5001"

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

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
	fmt.Println("====== Received Message ======")
	fmt.Printf("Command Type: %T\n", msg.cmd)
	fmt.Printf("Peer Authenticated: %v\n", msg.peer.authenticated)

	// Allow AUTH command without authentication
	if _, isAuth := msg.cmd.(AuthCommand); !isAuth {
		if !msg.peer.authenticated {
			fmt.Println("Authentication required, rejecting command")
			msg.peer.conn.Write([]byte("-NOAUTH Authentication required - AUTH {USERNAME} {PASSWORD}\r\n"))
			return nil
		}
	}

	switch v := msg.cmd.(type) {
	case SetCommand:
		fmt.Printf("SET command - Key: %s, Value: %s\n", string(v.key), string(v.val))
		s.kv.Set(v.key, v.val)
	case GetCommand:
		fmt.Printf("GET command - Key: %s\n", string(v.key))
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}
	case DeleteCommand:
		fmt.Printf("DELETE command - Key: %s\n", string(v.key))
		s.kv.Delete(v.key)

	case AuthCommand:
		fmt.Printf("AUTH command - Username: %s, Password: %s\n", v.username, v.password)
		// Verify credentials
		validAuth := false
		if v.username != "" {
			validAuth = (v.username == s.Username && v.password == s.Password)
		} else {
			validAuth = (v.password == s.Password)
		}

		if validAuth {
			msg.peer.authenticated = true
			fmt.Println("✓ Authentication successful")
			msg.peer.conn.Write([]byte("+USERNAME and PASSWORD are correct\r\n"))
		} else {
			msg.peer.authenticated = false
			fmt.Println("✗ Authentication failed")
			msg.peer.conn.Write([]byte("-USERNAME or PASSWORD are incorrect\r\n"))
		}
	}
	fmt.Println("===========================")
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
	fmt.Printf("\n🔗 New connection from: %s\n", conn.RemoteAddr())
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	fmt.Println("📖 Starting read loop for peer...")
	go peer.readLoop()

}

func main() {
	loadEnv()
	server := NewServer(Config{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	})
	go func() {
		log.Fatal(server.Start())
	}()

	time.Sleep(1000 * time.Second)

}
