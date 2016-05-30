package main

import (
	"errors"

	"github.com/go-steem/rpc"
)

type BlockMapReducer struct {
	endpointAddress string

	emitCh chan interface{}
	termCh chan struct{}
}

func NewBlockMapReducer(endpointAddress string) *BlockMapReducer {
	return &BlockMapReducer{
		endpointAddress: endpointAddress,
		emitCh:          make(chan interface{}, 0),
		termCh:          make(chan struct{}),
	}
}

func (reducer *BlockMapReducer) Run(startingBlock uint32) error {
	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return err
	}

}

func (reducer *BlockMapReducer) Stop() error {
	select {
	case <-reducer.termCh:
		return errors.New("terminated")
	default:
		close(reducer.termCh)
		return nil
	}
}

func (reducer *BlockMapReducer) emit(value interface{}) error {
	select {
	case reducer.emitCh <- value:
	case <-reducer.termCh:
		return errors.New("terminating")
	}
}
