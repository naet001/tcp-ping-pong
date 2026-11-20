package main

import (
	"bufio"
	"errors"
	"net"
	"testing"
	"time"
)

func TestPingTimesOutWhenServerDoesNotReply(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	holdConnection := make(chan struct{})
	defer close(holdConnection)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		if _, err := bufio.NewReader(conn).ReadString('\n'); err != nil {
			return
		}
		<-holdConnection
	}()

	start := time.Now()
	_, err = ping(listener.Addr().String(), 50*time.Millisecond)
	var timeoutError net.Error
	if !errors.As(err, &timeoutError) || !timeoutError.Timeout() {
		t.Fatalf("ping error = %v, want timeout", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("ping did not return promptly after its deadline")
	}
}
