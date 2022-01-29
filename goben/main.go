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

const version = "0.7"

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
	chart          string
	export         string
	csv            string
	ascii          bool // plot ascii chart
	tlsCert        string
	tlsKey         string
	tls            bool
	localAddr      string
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

func badExportFilename(parameter, filename string) error {
	if filename == "" {
		return nil
	}

	if strings.Contains(filename, "%d") && strings.Contains(filename, "%s") {
		return nil
	}

	return fmt.Errorf("badExportFilename %s: filename requires '%%d' and '%%s': %s", parameter, filename)
}

func main() {

	log.Printf("goben version " + version + " runtime " + runtime.Version() + " GOMAXPROCS=" + strconv.Itoa(runtime.GOMAXPROCS(0)) + " OS=" + runtime.GOOS + " arch=" + runtime.GOARCH)

	app := config{}

	flag.Var(&app.hosts, "hosts", "comma-separated list of hosts\nyou may append an optional port to every host: host[:port]")
	flag.Var(&app.listeners, "listeners", "comma-separated list of listen addresses\nyou may prepend an optional host to every port: [host]:port")
	flag.StringVar(&app.defaultPort, "defaultPort", ":8080", "default port")
	flag.IntVar(&app.connections, "connections", 1, "number of parallel connections")
	flag.StringVar(&app.reportInterval, "reportInterval", "2s", "periodic report interval\nunspecified time unit defaults to second")
	flag.StringVar(&app.totalDuration, "totalDuration", "10s", "test total duration\nunspecified time unit defaults to second")
	flag.IntVar(&app.opt.TCPReadSize, "tcpReadSize", 1000000, "TCP read buffer size in bytes")
	flag.IntVar(&app.opt.TCPWriteSize, "tcpWriteSize", 1000000, "TCP write buffer size in bytes")
	flag.IntVar(&app.opt.UDPReadSize, "udpReadSize", 64000, "UDP read buffer size in bytes")
	flag.IntVar(&app.opt.UDPWriteSize, "udpWriteSize", 64000, "UDP write buffer size in bytes")
	flag.BoolVar(&app.passiveClient, "passiveClient", false, "suppress client writes")
	flag.BoolVar(&app.opt.PassiveServer, "passiveServer", false, "suppress server writes")
	flag.Float64Var(&app.opt.MaxSpeed, "maxSpeed", 0, "bandwidth limit in mbps (0 means unlimited)")
	flag.BoolVar(&app.udp, "udp", false, "run client in UDP mode")
	flag.StringVar(&app.chart, "chart", "", "output filename for rendering chart on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -chart chart-%d-%s.png")
	flag.StringVar(&app.export, "export", "", "output filename for YAML exporting test results on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -export export-%d-%s.yaml")
	flag.StringVar(&app.csv, "csv", "", "output filename for CSV exporting test results on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -csv export-%d-%s.csv")
	flag.BoolVar(&app.ascii, "ascii", true, "plot ascii chart")
	flag.StringVar(&app.tlsKey, "key", "key.pem", "TLS key file")
	flag.StringVar(&app.tlsCert, "cert", "cert.pem", "TLS cert file")
	flag.BoolVar(&app.tls, "tls", true, "set to false to disable TLS")
	flag.StringVar(&app.localAddr, "localAddr", "", "bind specific local address:port\nexample: -localAddr 127.0.0.1:2000")

	flag.Parse()

	if errChart := badExportFilename("-chart", app.chart); errChart != nil {
		log.Panicf("%s", errChart.Error())
	}

	if errExport := badExportFilename("-export", app.export); errExport != nil {
		log.Panicf("%s", errExport.Error())
	}

	if errCsv := badExportFilename("-csv", app.csv); errCsv != nil {
		log.Panicf("%s", errCsv.Error())
	}

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
