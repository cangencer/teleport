package main

import (
	"flag"
	"teleport/client"
	"teleport/server"
)

func main() {
	isClient := flag.Bool("c", false, "is client")
	address := flag.String("a", "0.0.0.0:5000", "address to use")

	flag.Parse()

	if *isClient {
		client.Run(address)
	} else {
		server.Run(address)
	}
}

// func main() {
// 	key := []byte{254, 60, 60, 60}
// 	differentKey := []byte{60, 60, 60}
// 	value := []byte{150, 151, 152}
// 	kv.Set(key, value)
// 	result := kv.Get(differentKey)
// 	fmt.Println(result)
// }
