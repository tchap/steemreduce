package main

import (
	"flag"
	"os"
)

const (
	EnvironmentKeyRPCEndpoint = "STEEMREDUCE_RPC_ENDPOINT"
	EnvironmentKeyMapReduceID = "STEEMREDUCE_MAPREDUCE_ID"
)

type Config struct {
	RPCEndpointAddress string
	MapReduceID        string
}

func GetConfig() (*Config, error) {
	// Process environment variables.
	var (
		endpointAddress = os.Getenv(EnvironmentKeyRPCEndpoint)
		mapReduceID     = os.Getenv(EnvironmentKeyMapReduceID)
	)

	// Process command line flags.
	flagRPCEndpoint := flag.String(
		"rpc_endpoint", "ws://localhost:8090", "steemd RPC endpoint address")
	flagMapReduceID := flag.String(
		"mapreduce_id", "", "MapReduce implementation to run")
	flag.Parse()

	// Merge.
	if endpointAddress == "" {
		endpointAddress = *flagRPCEndpoint
	}
	if mapReduceID == "" {
		mapReduceID = *flagMapReduceID
	}

	// Return.
	return &Config{
		RPCEndpointAddress: endpointAddress,
		MapReduceID:        mapReduceID,
	}, nil
}
