#!/bin/bash

CGO_ENABLED=0 GOOS=$1 GOARCH=amd64 go build -o ./bin/donkey
