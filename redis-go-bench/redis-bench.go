package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const roundDuration = time.Second * 10

func main() {
	address := flag.String("a", "", "address to use")
	flag.Parse()

	fmt.Printf("Will connect to %s\n", *address)
	client := redis.NewClient(&redis.Options{
		Addr:     *address + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	client.Set("key", "012345678901234567890123456789012", 0)

	for i := 0; i < 2; i++ {
		fmt.Printf("Starting round %d...\n", i)
		endOfRound := time.Now().Add(roundDuration)
		completedRoundtrips := 0
		for time.Now().Before(endOfRound) {
			_, err := client.Get("key").Result()
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
