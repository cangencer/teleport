package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
	ctx := context.Background()
	fmt.Println("starting server")
	go func() { server(ctx, "127.0.0.1:7777") }()
	fmt.Println("starting client")
	client(ctx, "127.0.0.1:7777", "bla bla")
}

const maxBufferSize = 1024

const timeout = 10 * time.Millisecond

const responsePrefix = "I got "

func client(ctx context.Context, address string, message string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	doneChan := make(chan error, 1)

	go func() {
		buffer := make([]byte, maxBufferSize)
		expectedResponse := responsePrefix + message
		for i := 0; i < 10; i++ {
			endOfRound := time.Now().Add(time.Second)
			completedRoundtrips := 0
			for time.Now().Before(endOfRound) {
				_, err := fmt.Fprint(conn, message)
				if err != nil {
					doneChan <- err
					return
				}
				deadline := time.Now().Add(timeout)
				err = conn.SetReadDeadline(deadline)
				if err != nil {
					doneChan <- err
					return
				}
				n, _, err := conn.ReadFrom(buffer)
				if err != nil {
					doneChan <- err
					return
				}
				response := string(buffer[:n])
				if response != expectedResponse {
					fmt.Printf("Wrong response, got '%s' instead of '%s'\n", response, expectedResponse)
				}
				completedRoundtrips++
			}
			meanRtt := 1000000.0 / float64(completedRoundtrips)
			fmt.Printf("Mean RTT was %f µs\n", meanRtt)
		}
		doneChan <- nil
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Client cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return
}

func server(ctx context.Context, address string) (err error) {
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
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
	}

	return
}
