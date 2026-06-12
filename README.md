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
- [Export](#export)

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
  -p, --defaultPort string      default port, automatically appended to hosts without explicit port (default ":8080")
  -e, --export strings          export mode: comma-separated or repeated flags of ascii, csv, yaml, png, or filenames
                                example: --export ascii,csv,result-%d-%s.yaml or -e my.yaml -e my.png
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
    2026/06/12 00:36:38 goben version 1.1.0 runtime go1.26.4 GOMAXPROCS=16 OS=linux arch=amd64
    2026/06/12 00:36:38 connections=1 defaultPort=:8080 listeners=[] hosts=[]
    2026/06/12 00:36:38 reportInterval=0s totalDuration=0s
    2026/06/12 00:36:38 server mode (use -hosts to switch to client mode)
    2026/06/12 00:36:38 listenTCP: TLS disabled
    2026/06/12 00:36:38 listenTCP: spawning TLS listener: :8080
    2026/06/12 00:36:38 listenUDP: UDP disabled

Client side:

    $ goben -H 127.0.0.1
    2026/06/12 00:36:39 goben version 1.1.0 runtime go1.26.4 GOMAXPROCS=16 OS=linux arch=amd64
    2026/06/12 00:36:39 connections=1 defaultPort=:8080 listeners=[] hosts=["127.0.0.1"]
    2026/06/12 00:36:39 reportInterval=0s totalDuration=0s
    2026/06/12 00:36:39 client mode, tcp protocol
    2026/06/12 00:36:39 open: opening TLS=false tcp 0/1: 127.0.0.1:8080
    2026/06/12 00:36:39 open: trying non-TLS TCP
    2026/06/12 00:36:39 handleConnectionClient: starting TCP 0/1 127.0.0.1:8080
    2026/06/12 00:36:39 handleConnectionClient: options sent: {1s 2s 1000000 1000000 64000 64000 false 0 map[]}
    2026/06/12 00:36:39 serverVersion=1.1.0
    2026/06/12 00:36:39 handleConnectionClient: TCP ack received
    2026/06/12 00:36:39 clientWriter: starting: 0/1 127.0.0.1:8080
    2026/06/12 00:36:39 clientReader: starting: 0/1 127.0.0.1:8080
    2026/06/12 00:36:40 0/1  report   clientReader rate: 21467.668538 Mbps   3503 rcv/s
    2026/06/12 00:36:40 0/1  report   clientWriter rate: 21837.537818 Mbps   2729 snd/s
    2026/06/12 00:36:41 handleConnectionClient: 2s timer
    2026/06/12 00:36:41 127.0.0.1:8080 input:
     21468 ┼─────────────────────────╮
     19321 ┤                         ╰───────────╮
     17174 ┤                                     ╰───╮
     15027 ┤                                         ╰──╮
     12881 ┤                                            ╰───╮
     10734 ┤                                                ╰───╮
      8587 ┤                                                    ╰───╮
      6440 ┤                                                        ╰──╮
      4294 ┤                                                           ╰───╮
      2147 ┤                                                               ╰───╮
         0 ┤                                                                   ╰─
                       Input Mbps: 127.0.0.1:8080 Connection 0
    2026/06/12 00:36:41 127.0.0.1:8080 output:
     21838 ┼──────╮
     21704 ┤      ╰──────╮
     21570 ┤             ╰──────╮
     21437 ┤                    ╰──────╮
     21303 ┤                           ╰──────╮
     21170 ┤                                  ╰──────╮
     21036 ┤                                         ╰──────╮
     20903 ┤                                                ╰──────╮
     20769 ┤                                                       ╰──────╮
     20635 ┤                                                              ╰─────╮
     20502 ┤                                                                    ╰
                       Output Mbps: 127.0.0.1:8080 Connection 0
    2026/06/12 00:36:41 handleConnectionClient: closing: 0/1 127.0.0.1:8080
    2026/06/12 00:36:41 aggregate reading: 20748.578718 Mbps 3006 recv/s
    2026/06/12 00:36:41 aggregate writing: 21170.406833 Mbps 2646 send/s

# Export

Use `-e` / `--export` to save test results in one or more formats. Supported formats: `ascii`, `csv`, `yaml`, `png`.

Multiple formats can be combined with commas or repeated flags:

    # export CSV and YAML with auto-generated filenames
    goben -H 1.1.1.1 -e csv,yaml

    # export to specific filenames (extension determines format)
    goben -H 1.1.1.1 -e result.csv -e report.yaml

    # short form with multiple flags
    goben -H 1.1.1.1 -e ascii -e png

Without `--export`, ASCII charts are printed to the console only. Passing `--export ascii` also writes the chart to a file.

Auto-generated filenames use the pattern `result-<connIndex>-<host>.<ext>` (e.g. `result-0-127.0.0.1.csv`).

# TLS

For full a TLS setup please generate (all in PEM format):
- a self-signed CA (specify -CA CLI option on server and client)
  - you can have separate server and client CAs, for testing this is the same
- a server certificate and key (specify -cert and -key CLI options)
- a client certificate and key (specify -cert and -key CLI options)

You can use the minimal setup in the "certs" make target or look at the folder test/certs
for a simple local testing setup that might be adapted for all kinds of use cases.

--x--

