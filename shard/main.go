package main

import (
	"time"

	shard "github.com/evshiron/shitama/shard/lib"
)

func main() {

	shard := shard.NewShard()

	go shard.Start()

	for {
		time.Sleep(1 * time.Minute)
	}

}
