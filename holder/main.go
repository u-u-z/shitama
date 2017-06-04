package main

import (
	"time"

	holder "github.com/evshiron/shitama/holder/lib"
)

func main() {

	holder := holder.NewHolder()

	go holder.Start()

	for {
		time.Sleep(1 * time.Minute)
	}

}
