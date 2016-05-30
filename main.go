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
	// Process command line flags.
	flagStartingBlock := flag.Uint(
		"starting_block", 0, "block number to start with")
	flagSourceDirectory := flag.String(
		"source_directory", ".", "directory containing your MapReduce implementation")
	flag.Parse()

	var (
		startingBlock   = uint32(*flagStartingBlock)
		sourceDirectory = *flagSourceDirectory
	)
}
