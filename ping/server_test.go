package main

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"
)

func TestServeDrainsActiveConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan struct{})
	go func() {
		serve(ctx, listener, time.Second)
		close(serverDone)
	}()

	conn := dialServer(t, listener.Addr().String())
	defer conn.Close()
	response := bufio.NewReader(conn)
	writePing(t, conn)
	readPong(t, conn, response)

	if _, err := conn.Write([]byte("PI")); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := conn.Write([]byte("NG\n")); err != nil {
		t.Fatal(err)
	}
	readPong(t, conn, response)

	select {
	case <-serverDone:
	case <-time.After(time.Second):
		t.Fatal("server did not exit after the active request finished")
	}

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := response.ReadString('\n'); err == nil {
		t.Fatal("expected the server to close the connection after draining the request")
	}
}

func TestServeForceClosesConnectionsAfterGracePeriod(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan struct{})
	go func() {
		serve(ctx, listener, 50*time.Millisecond)
		close(serverDone)
	}()

	conn := dialServer(t, listener.Addr().String())
	response := bufio.NewReader(conn)
	writePing(t, conn)
	readPong(t, conn, response)

	cancel()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := response.ReadString('\n'); err == nil {
		t.Fatal("expected the server to close the idle connection after the grace period")
	}

	select {
	case <-serverDone:
	case <-time.After(time.Second):
		t.Fatal("server did not exit after closing the idle connection")
	}
	conn.Close()
}

func dialServer(t *testing.T, address string) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func writePing(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write([]byte("PING\n")); err != nil {
		t.Fatal(err)
	}
}

func readPong(t *testing.T, conn net.Conn, response *bufio.Reader) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	line, err := response.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != "PONG\n" {
		t.Fatalf("response = %q, want %q", line, "PONG\n")
	}
}
