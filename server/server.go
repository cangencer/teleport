package server

import (
	"context"
	"fmt"
	"time"

	"github.com/tidwall/evio"
)

// Run the server
func Run(localAddress *string) {
	ctx := context.Background()
	fmt.Printf("starting server on %s\n", *localAddress)
	server(ctx, localAddress)
}

const maxBufferSize = 1024
const timeout = time.Minute
const responsePrefix = "I got "

func server(ctx context.Context, localAddress *string) (err error) {
	var events evio.Events
	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		request := string(in[:])
		response := responsePrefix + request
		out = []byte(response)
		return
	}
	return evio.Serve(events, "udp://"+*localAddress)
}
