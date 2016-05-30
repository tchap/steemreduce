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

	// Start catching signals.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Start MapReduce.
	reducer := NewBlockMapReducer(endpointAddress)
	ctx, err := reducer.Start(startingBlock)
	if err != nil {
		return err
	}

	// Interrupt the process when a signal is received.
	go func() {
		<-signalCh
		signal.Stop(signalCh)
		ctx.Interrupt()
	}()

	// Wait.
	return ctx.Wait()
}
