package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

const requestTimeout = 5 * time.Second

func main() {
	response, err := ping("localhost:9000", requestTimeout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server:", response)
}

func ping(address string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	conn, err := (&net.Dialer{Deadline: deadline}).Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := conn.SetDeadline(deadline); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintln(conn, "PING"); err != nil {
		return "", err
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return response, nil
}
