package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type hostList []string

type config struct {
	hosts      hostList
	listenAddr string
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

	flag.Var(&app.hosts, "hosts", "comma-separated list of host[:port]")
	flag.StringVar(&app.listenAddr, "listen", ":80", "listen address host:port")

	flag.Parse()

	log.Printf("listen=%s hosts=%q", app.listenAddr, app.hosts)

	if len(app.hosts) == 0 {
		log.Printf("server mode")
		return
	}

	log.Printf("client mode")
}
