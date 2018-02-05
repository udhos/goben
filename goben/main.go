package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

const version = "0.0"

type hostList []string

type config struct {
	hosts       hostList
	listeners   hostList
	defaultPort string
	connections int
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

	flag.Var(&app.hosts, "hosts", "comma-separated list of hosts -- host[:port]")
	flag.Var(&app.listeners, "listeners", "comma-separated list of listen addresses -- host:port")
	flag.StringVar(&app.defaultPort, "defaultPort", ":8080", "default port")
	flag.IntVar(&app.connections, "connections", 1, "number of parallel connections")

	flag.Parse()

	if len(app.listeners) == 0 {
		app.listeners = []string{app.defaultPort}
	}

	log.Printf("goben version " + version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)))

	log.Printf("connections=%d defaultPort=%s listeners=%q hosts=%q",
		app.connections, app.defaultPort, app.listeners, app.hosts)

	if len(app.hosts) == 0 {
		log.Printf("server mode (use -hosts to switch to client mode)")
		serve(&app)
		return
	}

	log.Printf("client mode")
	open(&app)
}
