package main

import (
	"flag"
)

func main() {
	if err := _main(); err != nil {
		log.Faralln("Error:", err)
	}
}

func _main() error {
	flagStartingBlock := flag.Uint("starting_block", 0, "block number to start with")
	flag.Parse()

	var (
		startingBlock = uint32(*flagStartingBlock)
	)

}
