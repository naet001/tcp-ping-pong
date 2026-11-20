package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	maxConnections      = 100
	readTimeout         = 30 * time.Second
	shutdownGracePeriod = 5 * time.Second
)

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Server listening on port 9000")
	serve(ctx, listener, shutdownGracePeriod)
}

func serve(ctx context.Context, listener net.Listener, gracePeriod time.Duration) {
	connections := make(map[net.Conn]struct{})
	var connectionsMu sync.Mutex
	var handlers sync.WaitGroup
	sem := make(chan struct{}, maxConnections)

	stopClosingListener := context.AfterFunc(ctx, func() {
		listener.Close()
	})
	defer stopClosingListener()

acceptLoop:
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Println("Accept error:", err)
			continue
		}

		select {
		case <-ctx.Done():
			conn.Close()
			break acceptLoop
		default:
		}

		select {
		case sem <- struct{}{}:
			connectionsMu.Lock()
			connections[conn] = struct{}{}
			connectionsMu.Unlock()

			handlers.Add(1)
			go func() {
				defer handlers.Done()
				defer func() { <-sem }()
				defer func() {
					connectionsMu.Lock()
					delete(connections, conn)
					connectionsMu.Unlock()
				}()
				handleConnection(ctx, conn)
			}()
		default:
			conn.Close()
			log.Println("Rejecting connection: server at capacity")
		}
	}

	drained := make(chan struct{})
	go func() {
		handlers.Wait()
		close(drained)
	}()

	timer := time.NewTimer(gracePeriod)
	defer timer.Stop()

	select {
	case <-drained:
		return
	case <-timer.C:
	}

	connectionsMu.Lock()
	for conn := range connections {
		conn.Close()
	}
	connectionsMu.Unlock()
	<-drained
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	log.Printf("Connection from %v", conn.RemoteAddr())
	defer log.Printf("Connection closed: %v", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)

	for {
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

		select {
		case <-ctx.Done():
			log.Printf("Connection %v: shutdown requested", conn.RemoteAddr())
			return
		default:
		}
	}
}
