package main

import (
	"time"

	client "github.com/evshiron/shitama/client/lib"
)

func main() {

	client := client.NewClient()

	go client.Start()

	for {
		time.Sleep(1 * time.Minute)
	}

}
