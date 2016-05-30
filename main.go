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
	flagRPCEndpoint = flag.String(
		"rpc_endpoint", "ws://localhost:8090", "steemd RPC endpoint address")
	flagStartingBlock := flag.Uint(
		"starting_block", 0, "block number to start with")
	flag.Parse()

	var (
		rpcEndpoint   = *flagRPCEndpoint
		startingBlock = uint32(*flagStartingBlock)
	)

	// Run the whole thing.
	return NewSteemReducer(rpcEndpoint).Run(startingBlock)
}
