package client

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Run the client
func Run(remoteAddress *string) {
	ctx := context.Background()
	fmt.Printf("starting client to %s\n", *remoteAddress)
	if err := client(ctx, remoteAddress); err != nil {
		fmt.Println(err)
	}
}

const maxBufferSize = 1024
const timeout = 10 * time.Millisecond
const message = "bla bla"
const responsePrefix = "I got "
const roundDuration = time.Second
const numRounds = 10

func client(ctx context.Context, address *string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", *address)
	if err != nil {
		fmt.Printf("Couldn't resolve address %s\n", *address)
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
		for i := 0; i < numRounds; i++ {
			endOfRound := time.Now().Add(roundDuration)
			completedRoundtrips := 0
			minRtt, maxRtt := 999999999, 0
			for start := time.Now(); start.Before(endOfRound); start = time.Now() {
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
				end := time.Now()
				if err != nil {
					doneChan <- err
					return
				}
				response := string(buffer[:n])
				if response != expectedResponse {
					fmt.Printf("Wrong response, got '%s' instead of '%s'\n", response, expectedResponse)
				}
				took := int(end.Sub(start))
				if took < minRtt {
					minRtt = took
				}
				if took > maxRtt {
					maxRtt = took
				}
				completedRoundtrips++
			}
			meanRtt := durationMicros(int(roundDuration)) / float64(completedRoundtrips)
			fmt.Printf("min %.1f µs max %.1f µs avg %.1f µs\n",
				durationMicros(minRtt),
				durationMicros(maxRtt),
				meanRtt)
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

func durationMicros(duration int) float64 {
	return float64(duration) / 1000
}
