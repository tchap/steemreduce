package main

import (
	"errors"

	"github.com/go-steem/rpc"
	"gopkg.in/tomb.v2"
)

type BlockMapReducer struct {
	endpointAddress string

	emitCh chan interface{}

	t tomb.Tomb
}

func NewBlockMapReducer(endpointAddress string) *BlockMapReducer {
	return &BlockMapReducer{
		endpointAddress: endpointAddress,
		emitCh:          make(chan interface{}, 0),
	}
}

func (reducer *BlockMapReducer) Run(startingBlock uint32) error {
	// XXX: Make sure we can call Run only once.

	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return err
	}

}

func (reducer *BlockMapReducer) Interrupt() {
	reducer.t.Kill(nil)
}

func (reducer *BlockMapReducer) Wait() error {
	reducer.t.Wait()
}

func (reducer *BlockMapReducer) loop() {
	client, err := rpc.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	// Get config.
	log.Println("---> GetConfig()")
	config, err := client.GetConfig()
	if err != nil {
		return err
	}

	// Use the last irreversible block number as the initial last block number.
	props, err := client.GetDynamicGlobalProperties()
	if err != nil {
		return err
	}
	lastBlock := props.LastIrreversibleBlockNum

	// Keep processing incoming blocks forever.
	log.Printf("---> Entering the block processing loop (last block = %v)\n", lastBlock)
	for {
		// Get current properties.
		props, err := client.GetDynamicGlobalProperties()
		if err != nil {
			return err
		}

		// Process new blocks.
		for props.LastIrreversibleBlockNum-lastBlock > 0 {
			block, err := client.GetBlock(lastBlock)
			if err != nil {
				return err
			}

			reducer.ProcessBlock()

			lastBlock++
		}

		// Sleep for STEEMIT_BLOCK_INTERVAL seconds before the next iteration.
		time.Sleep(time.Duration(config.SteemitBlockInterval) * time.Second)
	}
}

func (reducer *BlockMapReducer) emit(value interface{}) error {
	select {
	case reducer.emitCh <- value:
	case <-reducer.termCh:
		return errors.New("terminating")
	}
}
