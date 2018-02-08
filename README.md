[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/goben/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/goben)](https://goreportcard.com/report/github.com/udhos/goben)

# goben
goben is a golang tool to measure TCP/UDP transport layer throughput between hotsts.

Install
=======

    go get github.com/udhos/goben
    go install github.com/udhos/goben/goben

Usage
=====

Make sure ~/go/bin is in your shell PATH.

Start server:

    server$ goben

Start client:

    client$ goben -hosts 1.1.1.1 ;# 1.1.1.1 is server's address

Example
=======

Server side:

    $ goben
    2018/02/08 18:37:28 goben version 0.1 runtime go1.10rc2 GOMAXPROCS=1
    2018/02/08 18:37:28 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=[]
    2018/02/08 18:37:28 reportInterval=2s totalDuration=10s
    2018/02/08 18:37:28 server mode (use -hosts to switch to client mode)
    2018/02/08 18:37:28 serve: spawning TCP listener: :8080
    2018/02/08 18:37:28 serve: spawning UDP listener: :8080

Client side:

    $ goben -hosts localhost
    2018/02/08 18:38:48 goben version 0.1 runtime go1.10rc2 GOMAXPROCS=1
    2018/02/08 18:38:48 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=["localhost"]
    2018/02/08 18:38:48 reportInterval=2s totalDuration=10s
    2018/02/08 18:38:48 client mode, tcp protocol
    2018/02/08 18:38:48 open: opening tcp 0/1: localhost:8080
    2018/02/08 18:38:48 handleConnectionClient: starting 0/1 [::1]:8080
    2018/02/08 18:38:48 handleConnectionClient: options sent: {2s 10s 50000 50000 false 0}
    2018/02/08 18:38:48 clientReader: starting: 0/1 [::1]:8080
    2018/02/08 18:38:48 clientWriter: starting: 0/1 [::1]:8080
    2018/02/08 18:38:50 report   clientReader rate:  11565 Mbps  28913 rcv/s
    2018/02/08 18:38:50 report   clientWriter rate:  11189 Mbps  27973 snd/s
    2018/02/08 18:38:52 report   clientReader rate:  11340 Mbps  28352 rcv/s
    2018/02/08 18:38:52 report   clientWriter rate:  10975 Mbps  27438 snd/s
    2018/02/08 18:38:54 report   clientReader rate:  11647 Mbps  29117 rcv/s
    2018/02/08 18:38:54 report   clientWriter rate:  11272 Mbps  28180 snd/s
    2018/02/08 18:38:56 report   clientReader rate:  10957 Mbps  27394 rcv/s
    2018/02/08 18:38:56 report   clientWriter rate:  10603 Mbps  26508 snd/s
    2018/02/08 18:38:58 workLoop: clientWriter: write tcp [::1]:55186->[::1]:8080: write: connection reset by peer
    2018/02/08 18:38:58 average clientWriter rate: 10995 Mbps 27489 snd/s
    2018/02/08 18:38:58 clientWriter: exiting: 0/1 [::1]:8080
    2018/02/08 18:38:58 report   clientReader rate:  11297 Mbps  28244 rcv/s
    2018/02/08 18:38:58 handleConnectionClient: 10s timer
    2018/02/08 18:38:58 workLoop: clientReader: read tcp [::1]:55186->[::1]:8080: use of closed network connection
    2018/02/08 18:38:58 average clientReader rate: 11361 Mbps 28402 rcv/s
    2018/02/08 18:38:58 clientReader: exiting: 0/1 [::1]:8080
    2018/02/08 18:38:58 handleConnectionClient: closing: 0/1 [::1]:8080
