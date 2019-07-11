package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// GetCommand fetches the value for the given key
const GetCommand = "GET "

// SetCommand sets the given value for the given key
const SetCommand = "SET "

// OkStatus result OK
const OkStatus = "OK  "

// ErrorStatus error
const ErrorStatus = "ERR "

// Run the key-value store server
func Run(localAddress *string) {
	ctx := context.Background()
	fmt.Printf("starting server on %s\n", *localAddress)
	server(ctx, localAddress)
}

const maxBufferSize = 1024
const timeout = time.Minute
const cmdLen = 4
const sizeOfUint32 = 4

func server(ctx context.Context, address *string) (err error) {
	conn, err := net.ListenPacket("udp", *address)
	if err != nil {
		fmt.Printf("Couldn't resolve the local address %s\n", *address)
		return
	}
	defer conn.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		storage := make(map[string]string)
		for {
			requestLen, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}
			var response string
			if requestLen < cmdLen {
				response = ErrorStatus + "no command in request"
			}
			command := string(buffer[:cmdLen])
			switch command {
			case GetCommand:
				keyStart := cmdLen
				key := string(buffer[keyStart:requestLen])
				value := storage[key]
				response = OkStatus + string(value)
			case SetCommand:
				keyStart := cmdLen + sizeOfUint32
				keyLen := binary.BigEndian.Uint32(buffer[cmdLen:keyStart])
				valueStart := keyStart + int(keyLen)
				key := string(buffer[keyStart:valueStart])
				value := string(buffer[valueStart:requestLen])
				storage[key] = value
				response = OkStatus
			default:
				response = fmt.Sprintf(ErrorStatus+"invalid command '%s'", command)
			}
			requestLen = copy(buffer, response)
			requestLen, err = conn.WriteTo(buffer[:requestLen], addr)
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
