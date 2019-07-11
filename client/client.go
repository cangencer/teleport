package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"teleport/server"
	"time"
)

// Run the client
func Run(remoteAddress *string) {
	ctx := context.Background()
	fmt.Printf("starting client to %s\n", *remoteAddress)
	err := client(ctx, remoteAddress)
	if err != nil {
		fmt.Printf("Error in client: %s\n", err)
	}
}

const maxBufferSize = 1024
const timeout = 10 * time.Millisecond
const roundDuration = time.Second

type connState struct {
	conn   *net.UDPConn
	buffer []byte
}

func client(ctx context.Context, address *string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", *address)
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
		connState := connState{conn, make([]byte, maxBufferSize)}
		key := "key1"
		value := "value1"
		err := executeSet(connState, key, value)
		if err != nil {
			doneChan <- err
			return
		}
		for i := 0; i < 10; i++ {
			endOfRound := time.Now().Add(roundDuration)
			completedRoundtrips := 0
			for time.Now().Before(endOfRound) {
				gotValue, err := executeGet(connState, key)
				if err != nil {
					doneChan <- err
					return
				}
				if gotValue != value {
					doneChan <- fmt.Errorf("got unexpected value '%s'", gotValue)
					return
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

func executeGet(connState connState, key string) (value string, err error) {
	response, err := executeCommand(connState, server.GetCommand+key)
	if err != nil {
		return
	}
	status := response[:4]
	if status != server.OkStatus {
		err = errors.New("error")
		return
	}
	value = response[4:]
	return
}

func executeSet(connState connState, key string, value string) error {
	var keyLenEncoded [4]byte
	binary.BigEndian.PutUint32(keyLenEncoded[:], uint32(len(key)))
	response, err := executeCommand(connState, server.SetCommand+string(keyLenEncoded[:])+key+value)
	if err != nil {
		return err
	}
	status := response[:4]
	if status != server.OkStatus {
		return errors.New(response)
	}
	return nil
}

func executeCommand(connState connState, request string) (response string, err error) {
	conn := connState.conn
	buffer := connState.buffer
	_, err = fmt.Fprint(connState.conn, request)
	if err != nil {
		return
	}
	deadline := time.Now().Add(timeout)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return
	}
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return
	}
	response = string(buffer[:n])
	return
}
