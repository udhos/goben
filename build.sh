#!/bin/sh

#go get github.com/wcharczuk/go-chart
#go get gopkg.in/v2/yaml
#go get github.com/guptarohit/asciigraph

gobin=~/go/bin

gofmt -s -w ./goben
go tool fix ./goben
go tool vet ./goben

#[ -x $gobin/gosimple ] && $gobin/gosimple goben/*.go
[ -x $gobin/golint ] && $gobin/golint goben/*.go
#[ -x $gobin/staticcheck ] && $gobin/staticcheck goben/*.go

go test ./goben
go install -v ./goben
