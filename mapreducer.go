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

func (reducer *BlockMapReducer) emit(value interface{}) error {
	select {
	case reducer.emitCh <- value:
	case <-reducer.termCh:
		return errors.New("terminating")
	}
}
