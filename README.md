[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/goben/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/goben)](https://goreportcard.com/report/github.com/udhos/goben)
[![GolangCI](https://golangci.com/badges/github.com/udhos/goben.svg)](https://golangci.com)

[![asciicast](https://asciinema.org/a/lnkGSPBNIY8o3jhTtBJRLHLrh.png)](https://asciinema.org/a/lnkGSPBNIY8o3jhTtBJRLHLrh)

# goben
goben is a golang tool to measure TCP/UDP transport layer throughput between hosts.

* [Features](#features)
* [History](#history)
* [Requirements](#requirements)
* [Install](#install)
* [Usage](#usage)
* [Example](#example)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

Features
========

- Support for TCP and UDP.
- Can limit maximum bandwidth.
- Written in [Go](https://golang.org/). Single executable file. No runtime dependency.
- Simple usage: start the server then launch the client pointing to server's address.
- Spawns multiple concurrent lightweight goroutines to handle multiple parallel traffic streams.
- Can save test results as PNG chart.
- Can export test results as YAML or CSV.

History
=======

- Years ago out of frustration with [iperf2](https://sourceforge.net/projects/iperf2/) limitations, I wrote the [nepim](http://www.nongnu.org/nepim/) tool. One can find some known iperf problems here: [iperf caveats](https://support.cumulusnetworks.com/hc/en-us/articles/216509388-Throughput-Testing-and-Troubleshooting#network_testing_with_open_source_tools). Nepim was more customizable, easier to use, reported simpler to understand results, was lighter on CPU.
- Later I found another amazing tool called [nuttcp](https://www.nuttcp.net/). One can read about nepim and nuttcp here: [nepim and nuttcp](https://www.linux.com/news/benchmarking-network-performance-network-pipemeter-lmbench-and-nuttcp).
- [goben](https://github.com/udhos/goben) is intended to fix shortcomings of nepim: (1) Take advantage of multiple CPUs while not wasting processing power. Nepim was single-threaded. (2) Be easily portable to multiple platforms. Nepim was heavily tied to UNIX-like world. (3) Use a simpler synchronous code flow. Nepim used hard-to-follow asynchronous architecture.

Requirements
============

- You need a [system with the Go language](https://golang.org/dl/) in order to build the application. There is no special requirement for running it.
- You can also download a binary release from https://github.com/udhos/goben/releases

Install
=======

## With Go Modules (since Go 1.11)

    git clone https://github.com/udhos/goben ;# clone outside GOPATH
    cd goben
    go test ./goben
    go install ./goben

## Without Go Modules (before Go 1.11)

    go get github.com/wcharczuk/go-chart
    go get gopkg.in/yaml.v2
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


--x--

