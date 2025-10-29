package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/resp"
)

const (
	CommandSet    = "SET"
	CommandGet    = "GET"
	CommandAuth   = "AUTH"
	CommandDelete = "DELETE"
)

type Command interface{}

type SetCommand struct {
	key []byte
	val []byte
}
type GetCommand struct {
	key []byte
}
type DeleteCommand struct {
	key []byte
}
type AuthCommand struct {
	username string
	password string
}

func parseCommand(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if v.Type() == resp.Array {
			for _, value := range v.Array() {
				switch value.String() {
				case CommandSet:
					if len(v.Array()) != 3 {
						return nil, fmt.Errorf("invalid SET command")
					}
					cmd := SetCommand{
						key: v.Array()[1].Bytes(),
						val: v.Array()[2].Bytes(),
					}
					return cmd, nil
				case CommandGet:
					if len(v.Array()) != 2 {
						return nil, fmt.Errorf("invalid GET command")
					}
					cmd := GetCommand{
						key: v.Array()[1].Bytes(),
					}
					return cmd, nil
				case CommandAuth:
					// AUTH can be: AUTH password  OR  AUTH username password
					var cmd AuthCommand
					if len(v.Array()) == 2 {
						cmd = AuthCommand{password: v.Array()[1].String()}
					} else if len(v.Array()) == 3 {
						cmd = AuthCommand{
							username: v.Array()[1].String(),
							password: v.Array()[2].String(),
						}
					} else {
						// Send error response
						return nil, fmt.Errorf("invalid AUTH command")
					}
					return cmd, nil
				default:
					//default case handling

				}

			}
		}
	}
	return fmt.Errorf("unknown or invalid command %s", raw), nil
}
