package goben

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/pflag"
)

// Version is the current application version.
const Version = "1.1.0"

// HostList holds a list of hosts to connect to or listen on.
type HostList []string

var (
	reExportMode = regexp.MustCompile(`^(?i)(ascii|csv|yaml|png)$`)
	reExportExt  = regexp.MustCompile(`(?i)\.(csv|yaml|yml|png|ascii)$`)
)

// ExportTarget holds an export mode and its output filename.
type ExportTarget struct {
	Mode     string
	Filename string
}

func defaultExportFilename(mode string) string {
	return fmt.Sprintf("result-%%d-%%s.%s", mode)
}

func parseExport(items []string) ([]ExportTarget, error) {
	if len(items) == 0 {
		return []ExportTarget{{Mode: "ascii", Filename: ""}}, nil
	}

	targets := make([]ExportTarget, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		lower := strings.ToLower(item)

		if reExportMode.MatchString(lower) {
			mode := lower
			targets = append(targets, ExportTarget{
				Mode:     mode,
				Filename: defaultExportFilename(mode),
			})
			continue
		}

		extMatch := reExportExt.FindStringSubmatch(lower)
		if extMatch != nil {
			ext := strings.TrimPrefix(extMatch[0], ".")
			mode := ext
			if mode == "yml" {
				mode = "yaml"
			}
			targets = append(targets, ExportTarget{
				Mode:     mode,
				Filename: item,
			})
			continue
		}

		return nil, fmt.Errorf("unrecognized export item: %q (expected ascii, csv, yaml, png, or a filename ending in .csv, .yaml, .yml, .png, .ascii)", item)
	}

	if len(targets) == 0 {
		return nil, nil
	}
	return targets, nil
}

// Config holds the configuration for the client and server.
type Config struct {
	Hosts          HostList
	Listeners      HostList
	DefaultPort    string
	Connections    int
	ReportInterval string
	TotalDuration  string
	Opt            Options
	PassiveClient  bool
	UDP            bool
	Export         []string
	exports        []ExportTarget
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
func (app *Config) AssignFlags(flagset *pflag.FlagSet) {
	flagset.VarP(&app.Hosts, "hosts", "H", "comma-separated list of target hosts for client mode\nformat: host[:port] (port defaults to --defaultPort)")
	flagset.VarP(&app.Listeners, "listeners", "l", "comma-separated list of listen addresses for server mode\nformat: [host]:port")
	flagset.StringVarP(&app.DefaultPort, "defaultPort", "p", ":8080", "default port, automatically appended to hosts without explicit port")
	flagset.IntVarP(&app.Connections, "connections", "c", 1, "number of parallel connections to each host")
	flagset.StringVarP(&app.ReportInterval, "reportInterval", "i", "2s", "periodic throughput report interval\nunspecified time unit defaults to second")
	flagset.StringVarP(&app.TotalDuration, "totalDuration", "d", "10s", "total test duration\nunspecified time unit defaults to second")
	flagset.IntVar(&app.Opt.TCPReadSize, "tcpReadSize", 1000000, "TCP read buffer size in bytes")
	flagset.IntVar(&app.Opt.TCPWriteSize, "tcpWriteSize", 1000000, "TCP write buffer size in bytes")
	flagset.IntVar(&app.Opt.UDPReadSize, "udpReadSize", 64000, "UDP read buffer size in bytes")
	flagset.IntVar(&app.Opt.UDPWriteSize, "udpWriteSize", 64000, "UDP write buffer size in bytes")
	flagset.BoolVar(&app.PassiveClient, "passiveClient", false, "suppress client traffic (receive only)")
	flagset.BoolVar(&app.Opt.PassiveServer, "passiveServer", false, "suppress server traffic (receive only)")
	flagset.Float64VarP(&app.Opt.MaxSpeed, "maxSpeed", "m", 0, "bandwidth limit in Mbps (0 means unlimited)")
	flagset.BoolVarP(&app.UDP, "udp", "u", false, "use UDP protocol instead of TCP")
	flagset.StringSliceVarP(&app.Export, "export", "e", nil, "export mode: comma-separated or repeated flags of ascii, csv, yaml, png, or filenames with recognized extensions\nexample: --export ascii,csv,result-%d-%s.yaml or -e my.yaml -e my.png")
	flagset.StringVar(&app.TLSKey, "key", "key.pem", "TLS private key file (PEM format)")
	flagset.StringVar(&app.TLSCert, "cert", "cert.pem", "TLS certificate file (PEM format)")
	flagset.StringVar(&app.TLSCA, "ca", "ca.pem", "TLS CA certificate file for peer verification (PEM format)")
	flagset.BoolVarP(&app.TLS, "tls", "s", true, "enable TLS encryption")
	flagset.BoolVar(&app.TLSAuthClient, "tlsAuthClient", true, "enable mutual TLS: verify server certificate against CA")
	flagset.BoolVar(&app.TLSAuthServer, "tlsAuthServer", true, "enable mutual TLS: verify client certificate against CA")
	flagset.BoolVarP(&app.TCP, "tcp", "t", true, "enable TCP transport (disable to test TLS-only or UDP-only)")
	flagset.StringVarP(&app.LocalAddr, "localAddr", "a", "", "bind specific local address:port\nexample: --localAddr 127.0.0.1:2000")
}

// NewDefaultConfig creates a config with default values
func NewDefaultConfig() *Config {
	result := Config{}
	result.Opt.Table = map[string]string{
		"clientVersion": Version,
	}

	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	result.AssignFlags(flagSet)
	_ = flagSet.Parse([]string{})
	return &result
}

func (h *HostList) String() string {
	return strings.Join(*h, ",")
}

// Set sets HostList value from a comma-separated list.
func (h *HostList) Set(value string) error {
	for hh := range strings.SplitSeq(value, ",") {
		*h = append(*h, hh)
	}
	return nil
}

// Type returns the type name for pflag display.
func (h *HostList) Type() string {
	return "strings"
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
	targets, errParse := parseExport(app.Export)
	if errParse != nil {
		return fmt.Errorf("invalid --export value: %w", errParse)
	}
	for _, t := range targets {
		if strings.Contains(t.Filename, "%") {
			if err := badExportFilename("--export", t.Filename); err != nil {
				log.Printf("%s", err.Error())
				return err
			}
		}
	}
	app.exports = targets

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
//
// Deprecated: Use ValidateAndUpdateConfig instead, which supersedes this function.
func ValidateAndUpdateServerConfig(app *Config) error {
	return ValidateAndUpdateConfig(app)
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
