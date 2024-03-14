#!/bin/bash
#! made for main node start for testing

# Remove the data directory if it exists, then create it anew
sudo rm -rf data && mkdir data

# Create each file with '0' as its initial content
echo "0" > data/batchCount.txt
echo "0" > data/blockCount.txt
echo "0" > data/transactionCount.txt

# start node
go run cmd/main.go start