package main

import (
	"errors"

	"github.com/go-steem/rpc"
	"gopkg.in/tomb.v2"
)

type BlockMapReducer struct {
	endpointAddress string
}

func NewBlockMapReducer(endpointAddress string) *BlockMapReducer {
	return &BlockMapReducer{endpointAddress}
}

func (reducer *BlockMapReducer) Start(startingBlock uint32) (*Context, error) {
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
	endingBlock := props.LastIrreversibleBlockNum

	// Start.
	return NewContext(client, startigBlock, endingBlock), nil
}
