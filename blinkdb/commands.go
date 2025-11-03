package blinkdb

const (
	// CommandSet is the SET command identifier
	CommandSet = "SET"
	// CommandGet is the GET command identifier
	CommandGet = "GET"
	// CommandAuth is the AUTH command identifier
	CommandAuth = "AUTH"
	// CommandDelete is the DELETE command identifier
	CommandDelete = "DELETE"
)

// Command represents a BlinkDB command interface
type Command interface{}

// SetCommand represents a SET command
type SetCommand struct {
	key []byte
	val []byte
}

// GetCommand represents a GET command
type GetCommand struct {
	key []byte
}

// DeleteCommand represents a DELETE command
type DeleteCommand struct {
	key []byte
}

// AuthCommand represents an AUTH command
type AuthCommand struct {
	username string
	password string
}
