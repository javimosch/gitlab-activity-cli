#!/bin/bash
set -e

echo "Building gitlab-activity-cli..."
go build -ldflags "-s -w" -o gitlab-activity-cli main.go
echo "Build complete: ./gitlab-activity-cli"