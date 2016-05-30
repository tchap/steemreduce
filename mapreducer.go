package main

import (
	"errors"

	"github.com/go-steem/rpc"
	"gopkg.in/tomb.v2"
)

type BlockMapReducer struct {
	endpointAddress string
	t               *tomb.Tomb
}

func NewBlockMapReducer(endpointAddress string) *BlockMapReducer {
	return &BlockMapReducer{
		endpointAddress: endpointAddress,
		emitCh:          make(chan interface{}, 0),
		t:               &tomb.Tomb{},
	}
}

// XXX: Make sure we can call Run only once.
func (reducer *BlockMapReducer) Run(startingBlock uint32) (*Context, error) {
	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return err
	}

	// Get the ending block number.
	props, err := client.GetDynamicGlobalProperties()
	if err != nil {
		return err
	}

	// Store what is needed later in the reducer.
	reducer.client = client
	reducer.startingBlock = startingBlock
	reducer.endingBlock = props.LastIrreversibleBlockNum

	// Start all the goroutines.
	reducer.t.Go(reducer.blockReader)
	reducer.t.Go(reducer.loop)
	return nil
}

func (reducer *BlockMapReducer) Interrupt() {
	reducer.t.Kill(nil)
}

func (reducer *BlockMapReducer) Wait() error {
	reducer.t.Wait()
}

func (reducer *BlockMapReducer) blockReader() {
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
