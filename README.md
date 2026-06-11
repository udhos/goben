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

    client$ goben -H 1.1.1.1 ;# 1.1.1.1 is server's address

# Command-line Options

Find several supported command-line switches by running 'goben -h':

```
$ goben -h
Usage of goben:
      --ca string               TLS CA certificate file for peer verification (PEM format) (default "ca.pem")
      --cert string             TLS certificate file (PEM format) (default "cert.pem")
  -c, --connections int         number of parallel connections to each host (default 1)
  -p, --defaultPort int         default port, automatically appended to hosts without explicit port (default 8080)
  -e, --export int              export mode: 1=ASCII (default), 2=CSV, 3=YAML, 4=PNG (default 1)
      --exportFile string       output filename for CSV/YAML/PNG export (supports %d=connIndex, %s=host)
  -H, --hosts strings           comma-separated list of target hosts for client mode
                                format: host[:port] (port defaults to --defaultPort)
      --key string              TLS private key file (PEM format) (default "key.pem")
  -l, --listeners strings       comma-separated list of listen addresses for server mode
                                format: [host]:port
  -a, --localAddr string        bind specific local address:port
                                example: --localAddr 127.0.0.1:2000
  -m, --maxSpeed float          bandwidth limit in Mbps (0 means unlimited)
      --passiveClient           suppress client traffic (receive only)
      --passiveServer           suppress server traffic (receive only)
  -i, --reportInterval string   periodic throughput report interval
                                unspecified time unit defaults to second (default "2s")
  -t, --tcp                     enable TCP transport (disable to test TLS-only or UDP-only) (default true)
      --tcpReadSize int         TCP read buffer size in bytes (default 1000000)
      --tcpWriteSize int        TCP write buffer size in bytes (default 1000000)
  -s, --tls                     enable TLS encryption (default true)
      --tlsAuthClient           enable mutual TLS: verify server certificate against CA (default true)
      --tlsAuthServer           enable mutual TLS: verify client certificate against CA (default true)
  -d, --totalDuration string    total test duration
                                unspecified time unit defaults to second (default "10s")
  -u, --udp                     use UDP protocol instead of TCP
      --udpReadSize int         UDP read buffer size in bytes (default 64000)
      --udpWriteSize int        UDP write buffer size in bytes (default 64000)
```

# Example

Server side:

    $ goben
    2026/06/10 01:49:33 goben version 1.0.3 runtime go1.26.2 GOMAXPROCS=16 OS=linux arch=amd64
    2026/06/10 01:49:33 connections=1 defaultPort=8080 listeners=[] hosts=[]
    2026/06/10 01:49:33 reportInterval=0s totalDuration=0s
    2026/06/10 01:49:33 server mode (use --hosts to switch to client mode)
    2026/06/10 01:49:33 listenTCP: TLS disabled
    2026/06/10 01:49:33 listenTCP: spawning TLS listener: :8080
    2026/06/10 01:49:33 listenUDP: UDP disabled

Client side:

    $ goben -H 127.0.0.1
    2026/06/10 01:49:34 goben version 1.0.3 runtime go1.26.2 GOMAXPROCS=16 OS=linux arch=amd64
    2026/06/10 01:49:34 connections=1 defaultPort=8080 listeners=[] hosts=["127.0.0.1"]
    2026/06/10 01:49:34 reportInterval=0s totalDuration=0s
    2026/06/10 01:49:34 client mode, tcp protocol
    2026/06/10 01:49:34 open: opening TLS=false tcp 0/1: 127.0.0.1:8080
    2026/06/10 01:49:34 open: trying non-TLS TCP
    2026/06/10 01:49:34 handleConnectionClient: starting TCP 0/1 127.0.0.1:8080
    2026/06/10 01:49:34 handleConnectionClient: options sent: {1s 2s 1000000 1000000 64000 64000 false 0 map[]}
    2026/06/10 01:49:34 serverVersion=1.0.3
    2026/06/10 01:49:34 handleConnectionClient: TCP ack received
    2026/06/10 01:49:34 clientWriter: starting: 0/1 127.0.0.1:8080
    2026/06/10 01:49:34 clientReader: starting: 0/1 127.0.0.1:8080
    2026/06/10 01:49:35 0/1  report   clientReader rate: 21994.914328 Mbps   3732 rcv/s
    2026/06/10 01:49:35 0/1  report   clientWriter rate: 21881.352904 Mbps   2735 snd/s
    2026/06/10 01:49:36 handleConnectionClient: 2s timer
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

