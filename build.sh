#!/bin/sh

go get github.com/wcharczuk/go-chart

gobin=~/go/bin

gofmt -s -w goben/*.go
go tool fix goben/*.go
go tool vet goben

[ -x $gobin/gosimple ] && $gobin/gosimple goben/*.go
[ -x $gobin/golint ] && $gobin/golint goben/*.go
[ -x $gobin/staticcheck ] && $gobin/staticcheck goben/*.go

go test github.com/udhos/goben/goben
go install -v github.com/udhos/goben/goben
