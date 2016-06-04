#!/bin/bash

mapreduce="$1"

export STEEMREDUCE_PARAMS_DATA_DIR="./data/$mapreduce"

exec ./steemreduce \
	-rpc_endpoint="ws://$(docker-machine ip default):8090" \
	-mapreduce_id="$mapreduce"
