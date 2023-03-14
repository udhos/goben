#!/bin/bash

go install golang.org/x/vuln/cmd/govulncheck@latest

gofmt -s -w .

revive ./...

go mod tidy

govulncheck ./...

export CGO_ENABLED=1

go test -race ./...

export CGO_ENABLED=0

go install ./...
