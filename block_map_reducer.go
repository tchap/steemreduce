package main

import (
	"github.com/go-steem/rpc"
)

type BlockMapReducer struct {
	endpointAddress string
}

func NewBlockMapReducer(endpointAddress string) *BlockMapReducer {
	return &BlockMapReducer{endpointAddress}
}

func (reducer *BlockMapReducer) Start(startingBlockNum uint32) (*Context, error) {
	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return nil, err
	}

	// Get the ending block number.
	props, err := client.GetDynamicGlobalProperties()
	if err != nil {
		return nil, err
	}
	endingBlockNum := props.LastIrreversibleBlockNum

	// Start.
	return NewContext(client, startingBlockNum, endingBlockNum), nil
}
