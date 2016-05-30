package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-steem/rpc"
)

func main() {
	if err := _main(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func _main() error {
	// Process command line flags.
	flagRPCEndpoint := flag.String(
		"rpc_endpoint", "ws://localhost:8090", "steemd RPC endpoint address")
	flagStartingBlock := flag.Uint(
		"starting_block", 0, "block number to start with")
	flagEndingBlock := flag.Uint(
		"ending_block", 0, "block number to end with")
	flag.Parse()

	var (
		endpointAddress  = *flagRPCEndpoint
		startingBlockNum = uint32(*flagStartingBlock)
		endingBlockNum   = uint32(*flagEndingBlock)
	)

	// Start catching signals.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Start MapReduce.
	ctx, err := start(endpointAddress, startingBlockNum, endingBlockNum)
	if err != nil {
		return err
	}

	// Interrupt the process when a signal is received.
	go func() {
		<-signalCh
		fmt.Println("---> Interrupt received, exiting ...")
		signal.Stop(signalCh)
		ctx.Interrupt()
	}()

	// Wait.
	return ctx.Wait()
}

func start(endpointAddress string, startingBlockNum, endingBlockNum uint32) (*Context, error) {
	// Get the RPC client.
	client, err := rpc.Dial(endpointAddress)
	if err != nil {
		return nil, err
	}

	// Get the ending block number if necessary.
	if endingBlockNum == 0 {
		props, err := client.GetDynamicGlobalProperties()
		if err != nil {
			return nil, err
		}
		endingBlockNum = props.LastIrreversibleBlockNum
	}

	// Start.
	return NewContext(client, startingBlockNum, endingBlockNum), nil
}
