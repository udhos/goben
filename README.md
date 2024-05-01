[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/goben/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/goben)](https://goreportcard.com/report/github.com/udhos/goben)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/goben.svg)](https://pkg.go.dev/github.com/udhos/goben)
[![GolangCI](https://golangci.com/badges/github.com/udhos/goben.svg)](https://golangci.com/r/github.com/udhos/goben)

# goben

goben is a golang tool to measure TCP/UDP transport layer throughput between hosts.

- [goben](#goben)
- [Features](#features)
- [History](#history)
- [Requirements](#requirements)
- [Install](#install)
  - [Install with Go Modules (since Go 1.11)](#install-with-go-modules-since-go-111)
  - [Run directly from source](#run-directly-from-source)
- [Usage](#usage)
- [Command-line Options](#command-line-options)
- [Example](#example)
- [TLS](#tls)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

# Features

- Support for TCP, UDP, TLS.
  - TLS can be enforced by disabling TCP and UDP
- Can limit maximum bandwidth.
- Written in [Go](https://golang.org/). Single executable file. No runtime dependency.
- Simple usage: start the server then launch the client pointing to server's address.
- Spawns multiple concurrent lightweight goroutines to handle multiple parallel traffic streams.
- Can save test results as PNG chart.
- Can export test results as YAML or CSV.

# History

- Years ago out of frustration with [iperf2](https://sourceforge.net/projects/iperf2/) limitations, I wrote the [nepim](http://www.nongnu.org/nepim/) tool. One can find some known iperf problems here: [iperf caveats](https://support.cumulusnetworks.com/hc/en-us/articles/216509388-Throughput-Testing-and-Troubleshooting#network_testing_with_open_source_tools). Nepim was more customizable, easier to use, reported simpler to understand results, was lighter on CPU.
- Later I found another amazing tool called [nuttcp](https://www.nuttcp.net/). One can read about nepim and nuttcp here: [nepim and nuttcp](https://www.linux.com/news/benchmarking-network-performance-network-pipemeter-lmbench-and-nuttcp).
- [goben](https://github.com/udhos/goben) is intended to fix shortcomings of nepim: (1) Take advantage of multiple CPUs while not wasting processing power. Nepim was single-threaded. (2) Be easily portable to multiple platforms. Nepim was heavily tied to UNIX-like world. (3) Use a simpler synchronous code flow. Nepim used hard-to-follow asynchronous architecture.

# Requirements

- You need a [system with the Go language](https://golang.org/dl/) in order to build the application. There is no special requirement for running it.
- You can also download a binary release from https://github.com/udhos/goben/releases

# Install

## Install with Go Modules (since Go 1.11)

    git clone https://github.com/udhos/goben ;# clone outside GOPATH
    cd goben
    go test -v ./...
    CGO_ENABLED=0 go install ./cmd/goben

## Run directly from source

    go run ./cmd/goben

# Usage

Make sure ~/go/bin is in your shell PATH and goben has been installed.

Start server:

    server$ goben

Start client:

    client$ goben -hosts 1.1.1.1 ;# 1.1.1.1 is server's address

# Command-line Options

Find several supported command-line switches by running 'goben -h':

```
$ goben -h
2021/02/28 00:43:28 goben version 0.6 runtime go1.16 GOMAXPROCS=12 OS=linux arch=amd64
Usage of goben:
  -ascii
        plot ascii chart (default true)
  -ca string
        TLS CA file (if server: CA to validate the client cert, if client: CA to validate the server cert) (default "ca.pem")
  -cert string
        TLS cert file (default "cert.pem")
  -chart string
        output filename for rendering chart on client
        '%d' is parallel connection index to host
        '%s' is hostname:port
        example: -chart chart-%d-%s.png
  -connections int
        number of parallel connections (default 1)
  -csv string
        output filename for CSV exporting test results on client
        '%d' is parallel connection index to host
        '%s' is hostname:port
        example: -csv export-%d-%s.csv
  -defaultPort string
        default port (default ":8080")
  -export string
        output filename for YAML exporting test results on client
        '%d' is parallel connection index to host
        '%s' is hostname:port
        example: -export export-%d-%s.yaml
  -hosts value
        comma-separated list of hosts
        you may append an optional port to every host: host[:port]
  -key string
        TLS key file (default "key.pem")
  -listeners value
        comma-separated list of listen addresses
        you may prepend an optional host to every port: [host]:port
  -localAddr string
        bind specific local address:port
        example: -localAddr 127.0.0.1:2000
  -maxSpeed float
        bandwidth limit in mbps (0 means unlimited)
  -passiveClient
        suppress client writes
  -passiveServer
        suppress server writes
  -reportInterval string
        periodic report interval
        unspecified time unit defaults to second (default "2s")
  -tcp
        set to false to disable TCP (this can be used to test TLS only or UDP only) (default true)
  -tcpReadSize int
        TCP read buffer size in bytes (default 1000000)
  -tcpWriteSize int
        TCP write buffer size in bytes (default 1000000)
  -tls
        set to false to disable TLS (default true)
  -tlsAuthClient
        set to true to enable client certificate authentication (check against CA) (default true)
  -tlsAuthServer
        set to true to enable server certificate authentication (check against CA) (default true)
  -totalDuration string
        test total duration
        unspecified time unit defaults to second (default "10s")
  -udp
        run client in UDP mode
  -udpReadSize int
        UDP read buffer size in bytes (default 64000)
  -udpWriteSize int
        UDP write buffer size in bytes (default 64000)
```

# Example

Server side:

    $ goben
    2018/06/28 15:04:26 goben version 0.3 runtime go1.11beta1 GOMAXPROCS=1
    2018/06/28 15:04:26 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=[]
    2018/06/28 15:04:26 reportInterval=2s totalDuration=10s
    2018/06/28 15:04:26 server mode (use -hosts to switch to client mode)
    2018/06/28 15:04:26 serve: spawning TCP listener: :8080
    2018/06/28 15:04:26 serve: spawning UDP listener: :8080

Client side:

    $ goben -hosts localhost
    2018/06/28 15:04:28 goben version 0.3 runtime go1.11beta1 GOMAXPROCS=1
    2018/06/28 15:04:28 connections=1 defaultPort=:8080 listeners=[":8080"] hosts=["localhost"]
    2018/06/28 15:04:28 reportInterval=2s totalDuration=10s
    2018/06/28 15:04:28 client mode, tcp protocol
    2018/06/28 15:04:28 open: opening tcp 0/1: localhost:8080
    2018/06/28 15:04:28 handleConnectionClient: starting 0/1 [::1]:8080
    2018/06/28 15:04:28 handleConnectionClient: options sent: {2s 10s 50000 50000 false 0}
    2018/06/28 15:04:28 clientReader: starting: 0/1 [::1]:8080
    2018/06/28 15:04:28 clientWriter: starting: 0/1 [::1]:8080
    2018/06/28 15:04:30 0/1  report   clientReader rate:  13917 Mbps  34793 rcv/s
    2018/06/28 15:04:30 0/1  report   clientWriter rate:  13468 Mbps  33670 snd/s
    2018/06/28 15:04:32 0/1  report   clientReader rate:  14044 Mbps  35111 rcv/s
    2018/06/28 15:04:32 0/1  report   clientWriter rate:  13591 Mbps  33978 snd/s
    2018/06/28 15:04:34 0/1  report   clientReader rate:  12934 Mbps  32337 rcv/s
    2018/06/28 15:04:34 0/1  report   clientWriter rate:  12517 Mbps  31294 snd/s
    2018/06/28 15:04:36 0/1  report   clientReader rate:  13307 Mbps  33269 rcv/s
    2018/06/28 15:04:36 0/1  report   clientWriter rate:  12878 Mbps  32196 snd/s
    2018/06/28 15:04:38 0/1  report   clientWriter rate:  13330 Mbps  33325 snd/s
    2018/06/28 15:04:38 0/1  report   clientReader rate:  13774 Mbps  34436 rcv/s
    2018/06/28 15:04:38 handleConnectionClient: 10s timer
    2018/06/28 15:04:38 workLoop: 0/1 clientWriter: write tcp [::1]:42130->[::1]:8080: use of closed network connection
    2018/06/28 15:04:38 0/1 average   clientWriter rate:  13157 Mbps  32892 snd/s
    2018/06/28 15:04:38 clientWriter: exiting: 0/1 [::1]:8080
    2018/06/28 15:04:38 workLoop: 0/1 clientReader: read tcp [::1]:42130->[::1]:8080: use of closed network connection
    2018/06/28 15:04:38 0/1 average   clientReader rate:  13595 Mbps  33989 rcv/s
    2018/06/28 15:04:38 clientReader: exiting: 0/1 [::1]:8080
    2018/06/28 15:04:38 input:
     14038 ┤          ╭────╮
     13939 ┤──────────╯    ╰╮
     13840 ┼                ╰─╮
     13741 ┤                  ╰╮                                    ╭──
     13641 ┤                   ╰╮                               ╭───╯
     13542 ┤                    ╰─╮                          ╭──╯
     13443 ┤                      ╰╮                     ╭───╯
     13344 ┤                       ╰─╮               ╭───╯
     13245 ┤                         ╰╮          ╭───╯
     13146 ┤                          ╰─╮    ╭───╯
     13047 ┤                            ╰────╯
     12948 ┤
    2018/06/28 15:04:38 output:
     13585 ┤          ╭────╮
     13489 ┤──────────╯    ╰╮
     13393 ┼                ╰─╮
     13297 ┤                  ╰╮                                    ╭──
     13201 ┤                   ╰╮                               ╭───╯
     13105 ┤                    ╰─╮                          ╭──╯
     13009 ┤                      ╰╮                     ╭───╯
     12914 ┤                       ╰─╮               ╭───╯
     12818 ┤                         ╰╮          ╭───╯
     12722 ┤                          ╰─╮    ╭───╯
     12626 ┤                            ╰────╯
     12530 ┤
    2018/06/28 15:04:38 handleConnectionClient: closing: 0/1 [::1]:8080

# TLS

For full a TLS setup please generate (all in PEM format):
- a self-signed CA (specify -CA CLI option on server and client)
  - you can have separate server and client CAs, for testing this is the same
- a server certificate and key (specify -cert and -key CLI options)
- a client certificate and key (specify -cert and -key CLI options)

You can use the minimal setup in the "certs" make target or look at the folder test/certs
for a simple local testing setup that might be adapted for all kinds of use cases.

--x--

