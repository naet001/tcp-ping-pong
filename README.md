# TCP Ping-Pong

A simple TCP-based ping-pong server and client implementation in Go.

## Overview

This project demonstrates basic TCP socket communication with a ping-pong pattern:
- The **server** listens on port 9000
- The **client** connects to the server and sends "PING"
- The **server** responds with "PONG"

## Project Structure

```
tcp-ping-pong/
├── ping/
│   ├── server.go    # TCP server implementation
│   ├── server       # Compiled server binary
│   └── go.mod       # Go module file
└── pong/
    └── client.go    # TCP client implementation
```

## Running the Project

### Start the Server

```bash
cd ping
go run server.go
# Or use the pre-compiled binary
./server
```

The server will start and display:
```
Server listening on port 9000
```

### Run the Client

```bash
cd pong
go run client.go
```

The client will connect to the server and display:
```
Server: PONG
```

## Testing

To test the communication:

1. In one terminal, start the server:
   ```bash
   cd ping && ./server
   ```

2. In another terminal, run the client:
   ```bash
   cd pong && go run client.go
   ```

You should see "Server: PONG" as the output.

## Requirements

- Go 1.20+

## License

MIT License
