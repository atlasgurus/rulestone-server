#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o build/rulestone_linux_amd64 ./main/main.go

docker build -t rulestone:latest .