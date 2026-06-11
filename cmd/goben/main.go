// Package main implements the tool.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/udhos/goben/goben"
)

func main() {

	app := goben.Config{}

	app.AssignFlags(pflag.CommandLine)
	pflag.Parse()

	log.Print("goben version " + goben.Version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)) + " OS=" + runtime.GOOS + " arch=" + runtime.GOARCH)
	log.Printf("connections=%d defaultPort=%s listeners=%q hosts=%q",
		app.Connections, app.DefaultPort, app.Listeners, app.Hosts)
	log.Printf("reportInterval=%s totalDuration=%s", app.Opt.ReportInterval, app.Opt.TotalDuration)

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, gracefully shutting down...", sig)
		cancel()
	}()

	if len(app.Hosts) == 0 {
		log.Printf("server mode (use -hosts to switch to client mode)")

		var wg sync.WaitGroup
		listenSuccess := goben.Serve(ctx, &app, &wg)
		if !listenSuccess {
			log.Println("server failed to listen")
			cancel()
			return
		}

		<-ctx.Done()
		wg.Wait() // Wait for listener goroutines to finish cleanup
		return
	}

	var proto string
	if app.UDP {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	log.Printf("client mode, %s protocol", proto)

	if _, err := goben.Open(ctx, &app); err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
}
