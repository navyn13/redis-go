# BlinkDB

A high-performance, Redis-protocol compatible key-value store implementation in Go.

## Features

- Redis protocol compatible
- Key-value storage with SET, GET, and DELETE operations
- Authentication support
- Thread-safe operations
- Easy integration as a Go package
- Minimal memory footprint
- Blazing fast performance

## Installation

```bash
go get github.com/navyn13/redis-go/blinkdb
```

## Usage

### As a Package

```go
package main

import (
    "log"
    "os"
    "github.com/navyn13/redis-go/blinkdb"
)

func main() {
    // Create a new server instance
    server := redis.NewServer(redis.Config{
        ListenAddr: ":6379",  // Optional, defaults to :5001
        Username: "myuser",   // Optional
        Password: "mypass",   // Optional
    })

    // Start the server
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Running as a Standalone Server

1. Clone the repository
2. Create a `.env` file with your credentials:
```
USERNAME=myuser
PASSWORD=mypass
```
3. Run the server:
```bash
go run main.go
```

## Client Connection

You can connect to the server using any Redis client. Example using the Redis CLI:

```bash
redis-cli -p 5001
```

Authentication:
```
AUTH myuser mypass
```

Basic operations:
```
SET mykey myvalue
GET mykey
DELETE mykey
```

## License

MIT License