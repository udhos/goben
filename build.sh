#!/bin/sh

gobin=~/go/bin

gofmt -s -w goben/*
go tool fix goben/*
go tool vet goben

lint() {
    local f=$1
    [ -x $gobin/gosimple ]    && $gobin/gosimple    $f
    [ -x $gobin/golint ]      && $gobin/golint      $f
    [ -x $gobin/staticcheck ] && $gobin/staticcheck $f
}

lint main.go
lint server.go
lint client.go

go test github.com/udhos/goben/goben
go install -v github.com/udhos/goben/goben
