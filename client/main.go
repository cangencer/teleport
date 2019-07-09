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
	remoteAddress := os.Args[1]
	fmt.Println("starting client")
	client(ctx, remoteAddress)
}

const maxBufferSize = 1024
const timeout = 10 * time.Millisecond
const message = "bla bla"
const responsePrefix = "I got "
const roundDuration = time.Second

func client(ctx context.Context, address string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Couldn't resolve address %s\n", address)
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
			endOfRound := time.Now().Add(roundDuration)
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
			meanRtt := float64(roundDuration) / float64(completedRoundtrips) / 1000
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
