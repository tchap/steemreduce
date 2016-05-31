package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tchap/steemreduce/runner"

	"github.com/go-steem/rpc"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, "\nError:", err)
		os.Exit(1)
	}
}

func _main() error {
	// Load configuration.
	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Start catching signals.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Start MapReduce.
	ctx, err := start(config)
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

func start(config *Config) (*runner.Context, error) {
	// Get the RPC client.
	client, err := rpc.Dial(config.RPCEndpointAddress)
	if err != nil {
		return nil, err
	}

	// Get the chosen MapReduce implementation.
	implementation, ok := availableMapReducers[config.MapReduceID]
	if !ok {
		fmt.Fprintf(os.Stderr, `
Unknown MapReduce implementation: "%v"

Available implementations:

`, config.MapReduceID)

		for _, id := range availableMapReducerIDs {
			fmt.Fprintln(os.Stderr, "    ", id)
		}

		return nil, errors.New("unknown MapReduce implementation")
	}

	// Start the beast.
	return runner.Run(client, implementation)
}
