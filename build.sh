#!/bin/sh

gobin=~/go/bin

gofmt -s -w goben/*
go tool fix goben/*
go tool vet goben
[ -x $gobin/gosimple ] && $gobin/gosimple main.go
[ -x $gobin/golint ] && $gobin/golint main.go
[ -x $gobin/staticcheck ] && $gobin/staticcheck main.go
go test github.com/udhos/goben/goben
go install -v github.com/udhos/goben/goben
