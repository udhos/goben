package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const version = "0.0"

type hostList []string

type config struct {
	hosts                  hostList
	listeners              hostList
	defaultPort            string
	connections            int
	reportInterval         string
	totalDuration          string
	durationReportInterval time.Duration
	durationTotalDuration  time.Duration
}

func (h *hostList) String() string {
	return fmt.Sprint(*h)
}

func (h *hostList) Set(value string) error {
	for _, hh := range strings.Split(value, ",") {
		log.Printf("cmd-line host: %s", hh)
		*h = append(*h, hh)
	}
	return nil
}

func main() {

	app := config{}

	flag.Var(&app.hosts, "hosts", "comma-separated list of hosts\nyou may append an optional port to every host: host[:port]")
	flag.Var(&app.listeners, "listeners", "comma-separated list of listen addresses\nyou may prepend an optional host to every port: [host]:port")
	flag.StringVar(&app.defaultPort, "defaultPort", ":8080", "default port")
	flag.IntVar(&app.connections, "connections", 1, "number of parallel connections")
	flag.StringVar(&app.reportInterval, "reportInterval", "2s", "periodic report interval")
	flag.StringVar(&app.totalDuration, "totalDuration", "10s", "test total duration")

	flag.Parse()

	var errInterval error
	app.durationReportInterval, errInterval = time.ParseDuration(app.reportInterval)
	if errInterval != nil {
		log.Panicf("bad reportInterval: %q: %v", app.reportInterval, errInterval)
	}

	var errDuration error
	app.durationTotalDuration, errDuration = time.ParseDuration(app.totalDuration)
	if errDuration != nil {
		log.Panicf("bad totalDuration: %q: %v", app.totalDuration, errDuration)
	}

	if len(app.listeners) == 0 {
		app.listeners = []string{app.defaultPort}
	}

	log.Printf("goben version " + version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)))
	log.Printf("connections=%d defaultPort=%s listeners=%q hosts=%q",
		app.connections, app.defaultPort, app.listeners, app.hosts)
	log.Printf("reportInterval=%s totalDuration=%s", app.durationReportInterval, app.durationTotalDuration)

	if len(app.hosts) == 0 {
		log.Printf("server mode (use -hosts to switch to client mode)")
		serve(&app)
		return
	}

	log.Printf("client mode")
	open(&app)
}
