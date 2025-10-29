package main

import (
	"fmt"
	"io"
	"net"

	"github.com/tidwall/resp"
)

type Peer struct {
	conn          net.Conn
	msgCh         chan Message
	authenticated bool
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)

}

func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if v.Type() == resp.Array {
			for _, value := range v.Array() {
				switch value.String() {
				case CommandSet:
					if len(v.Array()) != 3 {
						return fmt.Errorf("invalid number of vairbales SET command")
					}
					cmd := SetCommand{
						key: v.Array()[1].Bytes(),
						val: v.Array()[2].Bytes(),
					}
					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
				case CommandGet:
					if len(v.Array()) != 2 {
						return fmt.Errorf("invalid number of vairbales GET command")
					}
					cmd := GetCommand{
						key: v.Array()[1].Bytes(),
					}
					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
				case CommandAuth:
					var cmd AuthCommand
					if len(v.Array()) == 2 {
						cmd = AuthCommand{password: v.Array()[1].String()}
					} else if len(v.Array()) == 3 {
						cmd = AuthCommand{
							username: v.Array()[1].String(),
							password: v.Array()[2].String(),
						}
					} else {
						p.conn.Write([]byte("-ERR wrong number of arguments for 'auth' command\r\n"))
						continue
					}
					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
				default:
					//default case handling

				}
			}
		}
	}
	return fmt.Errorf("unknown or invalid command")
}
