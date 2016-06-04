#!/bin/bash

mapreduce="$1"

export STEEMREDUCE_PARAMS_DATA_DIR="./data/$mapreduce"

exec ./steemreduce -mapreduce_id="$mapreduce"
