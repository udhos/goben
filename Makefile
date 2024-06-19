bin/goben:
	go build -o $@ ./cmd/goben/

bin/amd64/goben:
	GOOS=linux GOARCH=amd64 go build -o $@ ./cmd/goben/

test: certs
	go test -v ./...

certs:
	make -C test/certs

client: certs
	go run ./cmd/goben/ -hosts 127.0.0.1:8443 -totalDuration 5s -passiveClient -ca test/certs/ca.crt -cert test/certs/client.crt -key test/certs/client.key

server: certs
	go run ./cmd/goben/ -tls -key test/certs/ca.key -cert test/certs/ca.crt -listeners localhost:8443 -ca test/certs/ca.crt -tcp false -udp false

.PHONY: test certs client server
