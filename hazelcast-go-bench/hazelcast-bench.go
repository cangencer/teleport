package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
)

const roundDuration = time.Second * 10

func main() {
	address := flag.String("a", "", "address to use")
	flag.Parse()

	config := hazelcast.NewConfig()
	config.NetworkConfig().AddAddress(*address + ":5701")
	client, err := hazelcast.NewClientWithConfig(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	myMap, _ := client.GetMap("map")
	myMap.Put("key", "012345678901234567890123456789012")

	for i := 0; i < 2; i++ {
		fmt.Printf("Starting round %d...\n", i)
		endOfRound := time.Now().Add(roundDuration)
		completedRoundtrips := 0
		for time.Now().Before(endOfRound) {
			_, err := myMap.Get("key")
			if err != nil {
				panic(err)
			}
			completedRoundtrips++
		}
		fmt.Printf("Number of gets: %d\n", completedRoundtrips)
		meanRtt := float64(roundDuration) / float64(completedRoundtrips) / 1000
		fmt.Printf("Mean RTT was %f µs\n", meanRtt)
	}
}
