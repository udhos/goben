package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const version = "0.2"

type hostList []string

type config struct {
	hosts          hostList
	listeners      hostList
	defaultPort    string
	connections    int
	reportInterval string
	totalDuration  string
	opt            options
	passiveClient  bool // suppress client send
	udp            bool
}

type options struct {
	ReportInterval time.Duration
	TotalDuration  time.Duration
	ReadSize       int
	WriteSize      int
	PassiveServer  bool    // suppress server send
	MaxSpeed       float64 // mbps
}

func (h *hostList) String() string {
	return fmt.Sprint(*h)
}

func (h *hostList) Set(value string) error {
	for _, hh := range strings.Split(value, ",") {
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
	flag.StringVar(&app.reportInterval, "reportInterval", "2s", "periodic report interval\nunspecified time unit defaults to second")
	flag.StringVar(&app.totalDuration, "totalDuration", "10s", "test total duration\nunspecified time unit defaults to second")
	flag.IntVar(&app.opt.ReadSize, "readSize", 50000, "read buffer size in bytes")
	flag.IntVar(&app.opt.WriteSize, "writeSize", 50000, "write buffer size in bytes")
	flag.BoolVar(&app.passiveClient, "passiveClient", false, "suppress client writes")
	flag.BoolVar(&app.opt.PassiveServer, "passiveServer", false, "suppress server writes")
	flag.Float64Var(&app.opt.MaxSpeed, "maxSpeed", 0, "bandwidth limit in mbps (0 means unlimited)")
	flag.BoolVar(&app.udp, "udp", false, "run client in UDP mode")

	flag.Parse()

	app.reportInterval = defaultTimeUnit(app.reportInterval)
	app.totalDuration = defaultTimeUnit(app.totalDuration)

	var errInterval error
	app.opt.ReportInterval, errInterval = time.ParseDuration(app.reportInterval)
	if errInterval != nil {
		log.Panicf("bad reportInterval: %q: %v", app.reportInterval, errInterval)
	}

	var errDuration error
	app.opt.TotalDuration, errDuration = time.ParseDuration(app.totalDuration)
	if errDuration != nil {
		log.Panicf("bad totalDuration: %q: %v", app.totalDuration, errDuration)
	}

	if len(app.listeners) == 0 {
		app.listeners = []string{app.defaultPort}
	}

	log.Printf("goben version " + version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)))
	log.Printf("connections=%d defaultPort=%s listeners=%q hosts=%q",
		app.connections, app.defaultPort, app.listeners, app.hosts)
	log.Printf("reportInterval=%s totalDuration=%s", app.opt.ReportInterval, app.opt.TotalDuration)

	if len(app.hosts) == 0 {
		log.Printf("server mode (use -hosts to switch to client mode)")
		serve(&app)
		return
	}

	var proto string
	if app.udp {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	log.Printf("client mode, %s protocol", proto)
	open(&app)
}

// append "s" (second) to time string
func defaultTimeUnit(s string) string {
	if len(s) < 1 {
		return s
	}
	if unicode.IsDigit(rune(s[len(s)-1])) {
		return s + "s"
	}
	return s
}
