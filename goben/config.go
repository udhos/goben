package goben

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"
)

// Version is the current application version.
const Version = "1.1.0"

// HostList holds a list of hosts to connect to or listen on.
type HostList []string

// Config holds the configuration for the client and server.
type Config struct {
	Hosts          HostList
	Listeners      HostList
	DefaultPort    string
	Connections    int
	ReportInterval string
	TotalDuration  string
	Opt            Options
	PassiveClient  bool // suppress client send
	UDP            bool
	Chart          string
	Export         string
	CSV            string
	ASCII          bool // plot ascii chart
	TLSCert        string
	TLSKey         string
	TLSCA          string
	TLS            bool
	TLSAuthClient  bool
	TLSAuthServer  bool
	TCP            bool
	LocalAddr      string
}

// AssignFlags parses command line flags.
func (app *Config) AssignFlags(flagset *flag.FlagSet) {
	flagset.Var(&app.Hosts, "hosts", "comma-separated list of hosts\nyou may append an optional port to every host: host[:port]")
	flagset.Var(&app.Listeners, "listeners", "comma-separated list of listen addresses\nyou may prepend an optional host to every port: [host]:port")
	flagset.StringVar(&app.DefaultPort, "defaultPort", ":8080", "default port")
	flagset.IntVar(&app.Connections, "connections", 1, "number of parallel connections")
	flagset.StringVar(&app.ReportInterval, "reportInterval", "2s", "periodic report interval\nunspecified time unit defaults to second")
	flagset.StringVar(&app.TotalDuration, "totalDuration", "10s", "test total duration\nunspecified time unit defaults to second")
	flagset.IntVar(&app.Opt.TCPReadSize, "tcpReadSize", 1000000, "TCP read buffer size in bytes")
	flagset.IntVar(&app.Opt.TCPWriteSize, "tcpWriteSize", 1000000, "TCP write buffer size in bytes")
	flagset.IntVar(&app.Opt.UDPReadSize, "udpReadSize", 64000, "UDP read buffer size in bytes")
	flagset.IntVar(&app.Opt.UDPWriteSize, "udpWriteSize", 64000, "UDP write buffer size in bytes")
	flagset.BoolVar(&app.PassiveClient, "passiveClient", false, "suppress client writes")
	flagset.BoolVar(&app.Opt.PassiveServer, "passiveServer", false, "suppress server writes")
	flagset.Float64Var(&app.Opt.MaxSpeed, "maxSpeed", 0, "bandwidth limit in mbps (0 means unlimited)")
	flagset.BoolVar(&app.UDP, "udp", false, "run client in UDP mode")
	flagset.StringVar(&app.Chart, "chart", "", "output filename for rendering chart on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -chart chart-%d-%s.png")
	flagset.StringVar(&app.Export, "export", "", "output filename for YAML exporting test results on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -export export-%d-%s.yaml")
	flagset.StringVar(&app.CSV, "csv", "", "output filename for CSV exporting test results on client\n'%d' is parallel connection index to host\n'%s' is hostname:port\nexample: -csv export-%d-%s.csv")
	flagset.BoolVar(&app.ASCII, "ascii", true, "plot ascii chart")
	flagset.StringVar(&app.TLSKey, "key", "key.pem", "TLS key file")
	flagset.StringVar(&app.TLSCert, "cert", "cert.pem", "TLS cert file")
	flagset.StringVar(&app.TLSCA, "ca", "ca.pem", "TLS CA file (if server: CA to validate the client cert, if client: CA to validate the server cert)")
	flagset.BoolVar(&app.TLS, "tls", true, "set to false to disable TLS")
	flagset.BoolVar(&app.TLSAuthClient, "tlsAuthClient", true, "set to true to enable client certificate authentication (check against CA)")
	flagset.BoolVar(&app.TLSAuthServer, "tlsAuthServer", true, "set to true to enable server certificate authentication (check against CA)")
	flagset.BoolVar(&app.TCP, "tcp", true, "set to false to disable TCP (this can be used to test TLS only or UDP only)")
	flagset.StringVar(&app.LocalAddr, "localAddr", "", "bind specific local address:port\nexample: -localAddr 127.0.0.1:2000")
}

// NewDefaultConfig creates a config with default values
func NewDefaultConfig() *Config {
	result := Config{}
	result.Opt.Table = map[string]string{
		"clientVersion": Version,
	}

	flagSet := flag.FlagSet{}
	result.AssignFlags(&flagSet)
	flagSet.Parse([]string{})
	return &result
}

func (h *HostList) String() string {
	return fmt.Sprint(*h)
}

// Set sets HostList value from a comma-separated list.
func (h *HostList) Set(value string) error {
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

// ValidateAndUpdateConfig validates and updates the config
// it will set internal values necessary for successful completion
func ValidateAndUpdateConfig(app *Config) error {
	if errChart := badExportFilename("-chart", app.Chart); errChart != nil {
		log.Printf("%s", errChart.Error())
		return errChart
	}

	if errExport := badExportFilename("-export", app.Export); errExport != nil {
		log.Printf("%s", errExport.Error())
		return errExport
	}

	if errCsv := badExportFilename("-csv", app.CSV); errCsv != nil {
		log.Printf("%s", errCsv.Error())
		return errCsv
	}

	app.ReportInterval = defaultTimeUnit(app.ReportInterval)
	app.TotalDuration = defaultTimeUnit(app.TotalDuration)

	var errInterval error
	app.Opt.ReportInterval, errInterval = time.ParseDuration(app.ReportInterval)
	if errInterval != nil {
		log.Printf("bad reportInterval: %q: %v", app.ReportInterval, errInterval)
		return errInterval
	}

	var errDuration error
	app.Opt.TotalDuration, errDuration = time.ParseDuration(app.TotalDuration)
	if errDuration != nil {
		log.Printf("bad totalDuration: %q: %v", app.TotalDuration, errDuration)
		return errDuration
	}

	if len(app.Listeners) == 0 {
		app.Listeners = []string{app.DefaultPort}
	}

	return nil
}

// ValidateAndUpdateServerConfig validates and updates the config.
// It will set internal values necessary for successful completion.
func ValidateAndUpdateServerConfig(app *Config) error {
	if len(app.Listeners) == 0 {
		app.Listeners = []string{app.DefaultPort}
	}

	return nil
}

// append "s" (second) to time string
func defaultTimeUnit(s string) string {
	if s == "" {
		return s
	}
	if unicode.IsDigit(rune(s[len(s)-1])) {
		// last rune is digit
		return s + "s"
	}
	return s
}
