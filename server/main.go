package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	localAddress := os.Args[1]
	fmt.Printf("starting server on %s\n", localAddress)
	server(ctx, localAddress)
}

const maxBufferSize = 1024
const timeout = time.Minute
const responsePrefix = "I got "

func server(ctx context.Context, address string) (err error) {
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		fmt.Printf("Couldn't resolve the local address %s\n", address)
		return
	}
	defer conn.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}
			request := string(buffer[:n])
			response := responsePrefix + request
			n = copy(buffer, response)
			n, err = conn.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				return
			}
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Println("Server cancelled")
		err = ctx.Err()
	case err = <-doneChan:
		fmt.Printf("Got error: %s\n", err)
	}
	return
}
