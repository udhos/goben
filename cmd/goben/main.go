// Package main implements the tool.
package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"
	"sync"

	"github.com/udhos/goben/goben"
)

func main() {

	app := goben.Config{}

	// use the global flagset
	app.AssignFlags(flag.CommandLine)
	flag.Parse()

	log.Printf("goben version " + goben.Version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)) + " OS=" + runtime.GOOS + " arch=" + runtime.GOARCH)
	log.Printf("connections=%d defaultPort=%s listeners=%q hosts=%q",
		app.Connections, app.DefaultPort, app.Listeners, app.Hosts)
	log.Printf("reportInterval=%s totalDuration=%s", app.Opt.ReportInterval, app.Opt.TotalDuration)

	if len(app.Hosts) == 0 {
		log.Printf("server mode (use -hosts to switch to client mode)")

		var wg sync.WaitGroup
		wg.Add(1)
		listenSuccess := goben.Serve(&app, &wg)
		if !listenSuccess {
			log.Println("server failed to listen")
		}

		// wait until complete
		wg.Wait()
		return
	}

	var proto string
	if app.UDP {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	log.Printf("client mode, %s protocol", proto)
	goben.Open(&app)
}
