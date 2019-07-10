package main

import (
	"flag"

	"./client"
	"./server"
)

func main() {
	isClient := flag.Bool("c", false, "is client")
	address := flag.String("a", "localhost:5000", "address to use")

	flag.Parse()

	if *isClient {
		client.Run(address)
	} else {
		server.Run(address)
	}
}
