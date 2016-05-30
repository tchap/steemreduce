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
		endpointAddress = *flagRPCEndpoint
		startingBlock   = uint32(*flagStartingBlock)
	)

	// Instantiate a BlockMapReducer.
	reducer := NewBlockMapReducer(endpointAddress)

	// Start catching signals.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		signal.Stop(signalCh)
		reducer.Interrupt()
	}()

	// Run.
	return reducer.Run(startingBlock).Wait()
}
