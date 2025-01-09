#!/bin/sh
echo "Testing..."
echo
go test -v

echo "Building..."
echo
go build -o har2hosts

echo "Running..."
echo
./har2hosts $1
