package main

import (
	"github.com/go-steem/rpc"
)

type BlockMapReducer struct {
	endpointAddress string
}

func NewBlockMapReducer(endpointAddress string) *SteemReducer {
	return &BlockMapReducer{endpointAddress}
}

func (reducer *BlockMapReducer) Run(startingBlock uint32) error {
	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return err
	}

}
