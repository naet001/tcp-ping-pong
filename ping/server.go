package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	maxConnections = 100
	readTimeout    = 30 * time.Second
)

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Server listening on port 9000")

	// Semaphore for connection limiting
	sem := make(chan struct{}, maxConnections)

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Shutting down...")
				return
			default:
				log.Println("Accept error:", err)
			}
			continue
		}

		// Block if at capacity
		select {
		case sem <- struct{}{}:
			go func() {
				defer func() { <-sem }()
				handleConnection(ctx, conn)
			}()
		case <-ctx.Done():
			conn.Close()
			log.Println("Rejecting connection: shutting down")
		default:
			conn.Close()
			log.Println("Rejecting connection: server at capacity")
		}
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	log.Printf("Connection from %v", conn.RemoteAddr())
	defer log.Printf("Connection closed: %v", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)

	for {
		// Set read deadline for each iteration
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Printf("Scanner error from %v: %v", conn.RemoteAddr(), err)
			}
			return
		}

		message := scanner.Text()

		if message == "PING" {
			fmt.Fprintln(conn, "PONG")
		} else {
			log.Printf("Unexpected message from %v: %q", conn.RemoteAddr(), message)
			fmt.Fprintln(conn, "ERROR: expected PING")
		}

		// Check if shutdown was requested
		select {
		case <-ctx.Done():
			log.Printf("Connection %v: shutdown requested", conn.RemoteAddr())
			return
		default:
		}
	}
}
