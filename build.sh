#!/bin/sh

#go get github.com/wcharczuk/go-chart
#go get gopkg.in/v2/yaml
#go get github.com/guptarohit/asciigraph

gobin=~/go/bin

gofmt -s -w ./goben
go tool fix ./goben
go tool vet ./goben

#which gosimple >/dev/null && gosimple ./goben
which golint >/dev/null && golint ./goben
#which staticcheck >/dev/null && staticcheck ./goben

go test ./goben
go install -v ./goben
