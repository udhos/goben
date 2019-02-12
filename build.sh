#!/bin/sh

gofmt -s -w ./goben
go tool fix ./goben
go vet ./goben

#which gosimple >/dev/null && gosimple ./goben
which golint >/dev/null && golint ./goben
#which staticcheck >/dev/null && staticcheck ./goben

go test ./goben
go install -v ./goben
