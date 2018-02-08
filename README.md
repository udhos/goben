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
    2018/02/07 21:39:10 goben version 0.0 runtime go1.10rc2 GOMAXPROCS=1
    2018/02/07 21:39:10 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=[]
    2018/02/07 21:39:10 reportInterval=2s totalDuration=10s
    2018/02/07 21:39:10 server mode (use -hosts to switch to client mode)
    2018/02/07 21:39:10 serve: spawning TCP listener: :8080

Client side:

    $ goben -hosts localhost
    2018/02/07 21:40:19 cmd-line host: localhost
    2018/02/07 21:40:19 goben version 0.0 runtime go1.10rc2 GOMAXPROCS=1
    2018/02/07 21:40:19 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=["localhost"]
    2018/02/07 21:40:19 reportInterval=2s totalDuration=10s
    2018/02/07 21:40:19 client mode
    2018/02/07 21:40:19 open: opening TCP 0/1: localhost:8080
    2018/02/07 21:40:19 handleConnectionClient: starting 0/1 [::1]:8080
    2018/02/07 21:40:19 clientReader: starting: 0/1 [::1]:8080
    2018/02/07 21:40:19 clientWriter: starting: 0/1 [::1]:8080
    2018/02/07 21:40:21 report clientWriter rate:   7982 Mbps  49892 calls/s
    2018/02/07 21:40:21 report clientReader rate:   8238 Mbps  51496 calls/s
    2018/02/07 21:40:23 report clientWriter rate:   7849 Mbps  49058 calls/s
    2018/02/07 21:40:23 report clientReader rate:   8092 Mbps  50582 calls/s
    2018/02/07 21:40:25 report clientWriter rate:   7767 Mbps  48549 calls/s
    2018/02/07 21:40:25 report clientReader rate:   8291 Mbps  51829 calls/s
    2018/02/07 21:40:27 report clientWriter rate:   7549 Mbps  47183 calls/s
    2018/02/07 21:40:27 report clientReader rate:   8129 Mbps  50815 calls/s
    2018/02/07 21:40:29 workLoop: clientWriter: write tcp [::1]:54942->[::1]:8080: write: connection reset by peer
    2018/02/07 21:40:29 average clientWriter rate: 7739 Mbps 48370 calls/s
    2018/02/07 21:40:29 clientWriter: exiting: 0/1 [::1]:8080
    2018/02/07 21:40:29 handleConnectionClient: 10s timer
    2018/02/07 21:40:29 workLoop: clientReader: read tcp [::1]:54942->[::1]:8080: use of closed network connection
    2018/02/07 21:40:29 average clientReader rate: 8173 Mbps 51089 calls/s
    2018/02/07 21:40:29 clientReader: exiting: 0/1 [::1]:8080
    2018/02/07 21:40:29 handleConnectionClient: closing: 0/1 [::1]:8080

