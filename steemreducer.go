package main

import (
	"github.com/go-steem/rpc"
)

type SteemReducer struct {
	endpointAddress string
}

func NewSteemReducer(endpointAddress string) *SteemReducer {
	return &SteemReducer{endpointAddress}
}

func (reducer *SteemReducer) Run(startingBlock uint32) error {
	// Get the RPC client.
	client, err := rpc.Dial(reducer.endpointAddress)
	if err != nil {
		return err
	}

}
