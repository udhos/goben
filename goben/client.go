package goben

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Client struct {
}

type ClientStats struct {
	TotalDuration time.Duration
	ReadMbps      float64
	WriteMbps     float64
	ReadBytes     int64
	WriteBytes    int64
}

// open a client with a config and perform a test
func Open(app *Config) (ClientStats, error) {

	// validate the config first
	if err := ValidateAndUpdateConfig(app); err != nil {
		return ClientStats{}, err
	}

	// then open the connection
	var proto string
	if app.UDP {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	var wg sync.WaitGroup

	var aggReader aggregate
	var aggWriter aggregate

	dialer := net.Dialer{}

	if app.LocalAddr != "" {
		if app.UDP {
			addr, err := net.ResolveUDPAddr(proto, app.LocalAddr)
			if err != nil {
				log.Printf("open: resolve %s localAddr=%s: %v", proto, app.LocalAddr, err)
			}
			dialer.LocalAddr = addr
		} else {
			addr, err := net.ResolveTCPAddr(proto, app.LocalAddr)
			if err != nil {
				log.Printf("open: resolve %s localAddr=%s: %v", proto, app.LocalAddr, err)
			}
			dialer.LocalAddr = addr
		}
		log.Printf("open: localAddr: %s", dialer.LocalAddr)
	}

	successfulConnections := 0
	for _, h := range app.Hosts {

		hh := appendPortIfMissing(h, app.DefaultPort)

		for i := 0; i < app.Connections; i++ {

			log.Printf("open: opening TLS=%v %s %d/%d: %s", app.TLS, proto, i, app.Connections, hh)

			if !app.UDP && app.TLS {
				// try TLS first
				log.Printf("open: trying TLS")
				conn, errDialTLS := tlsDial(dialer, proto, hh, app)
				if errDialTLS == nil {
					spawnClient(app, &wg, conn, i, app.Connections, true, &aggReader, &aggWriter)
					successfulConnections++
					continue
				}
				log.Printf("open: trying TLS: failure: %s: %s: %v", proto, hh, errDialTLS)
			}

			if !app.UDP && app.TCP {
				log.Printf("open: trying non-TLS TCP")
			} else if !app.UDP {
				log.Printf("open: all enabled options failed to connect, aborting")
				continue
			}

			// only try TCP or UDP if explicitly enabled
			if app.UDP || app.TCP {
				conn, errDial := dialer.Dial(proto, hh)
				if errDial != nil {
					log.Printf("open: dial %s: %s: %v", proto, hh, errDial)
					continue
				}
				spawnClient(app, &wg, conn, i, app.Connections, false, &aggReader, &aggWriter)
				successfulConnections++
			}
		}
	}

	if successfulConnections == 0 {
		log.Printf("open: no successful connections")
		return ClientStats{}, fmt.Errorf("open: no successful connections")
	}

	wg.Wait()

	log.Printf("aggregate reading: %f Mbps %d recv/s", aggReader.Mbps, aggReader.Cps)
	log.Printf("aggregate writing: %f Mbps %d send/s", aggWriter.Mbps, aggWriter.Cps)

	return ClientStats{
		TotalDuration: app.Opt.TotalDuration,
		ReadMbps:      aggReader.Mbps,
		WriteMbps:     aggWriter.Mbps,
		ReadBytes:     aggReader.Bytes,
		WriteBytes:    aggWriter.Bytes,
	}, nil
}

func spawnClient(app *Config, wg *sync.WaitGroup, conn net.Conn, c, connections int, isTLS bool, aggReader, aggWriter *aggregate) {
	wg.Add(1)
	go handleConnectionClient(app, wg, conn, c, connections, isTLS, aggReader, aggWriter)
}

func tlsDial(dialer net.Dialer, proto, h string, app *Config) (net.Conn, error) {

	// load client cert, if this is not provided, it will not be sent along with the connection
	clientCerts := []tls.Certificate{}
	cert, err := tls.LoadX509KeyPair(app.TLSCert, app.TLSKey)
	if err != nil {
		log.Printf("tlsDial: failure loading TLS key pair: %v, will connect without explicitly specified key/cert", err)
	} else {
		clientCerts = append(clientCerts, cert)
	}

	// by default use the system cert pool
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("tlsDial: failure loading system cert pool: %v", err)
		return nil, err
	}

	// if a server auth is enabled, enforce auth against the CA file
	if app.TLSCA != "" {
		caCert, err := os.ReadFile(app.TLSCA)
		if err != nil {
			log.Printf("tlsDial: failure reading CA cert %s: %v", app.TLSCA, err)
			return nil, err
		}

		// create a cert pool
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	// set the TLS config
	conf := &tls.Config{
		InsecureSkipVerify: !app.TLSAuthServer,
		RootCAs:            caCertPool,
		Certificates:       clientCerts,
	}

	// and dial
	conn, err := tls.DialWithDialer(&dialer, proto, h, conf)
	if err != nil {
		log.Printf("tlsDial: %s %s: %v", proto, h, err)
		return nil, err
	}

	log.Println("client: connected to: ", conn.RemoteAddr())
	state := conn.ConnectionState()
	log.Println("Client: Server certificates:")
	for _, v := range state.PeerCertificates {
		log.Print("- Subject: ", v.Subject)
		log.Print("  Issuer: ", v.Issuer)
		log.Print("  Expiration: ", v.NotAfter)
	}
	log.Println("client: handshake complete: ", state.HandshakeComplete)

	return conn, err
}

// ExportInfo records data for export
type ExportInfo struct {
	Input  ChartData
	Output ChartData
}

func sendOptions(app *Config, conn io.Writer) error {
	opt := app.Opt
	if app.UDP {
		var optBuf bytes.Buffer
		enc := gob.NewEncoder(&optBuf)
		if errOpt := enc.Encode(&opt); errOpt != nil {
			log.Printf("handleConnectionClient: UDP options failure: %v", errOpt)
			return errOpt
		}
		_, optWriteErr := conn.Write(optBuf.Bytes())
		if optWriteErr != nil {
			log.Printf("handleConnectionClient: UDP options write: %v", optWriteErr)
			return optWriteErr
		}
	} else {
		enc := gob.NewEncoder(conn)
		if errOpt := enc.Encode(&opt); errOpt != nil {
			log.Printf("handleConnectionClient: TCP options failure: %v", errOpt)
			return errOpt
		}
	}
	return nil
}

func handleConnectionClient(app *Config, wg *sync.WaitGroup, conn net.Conn, c, connections int, isTLS bool, aggReader, aggWriter *aggregate) {
	defer wg.Done()

	log.Printf("handleConnectionClient: starting %s %d/%d %v", protoLabel(isTLS), c, connections, conn.RemoteAddr())

	// send options
	if errOpt := sendOptions(app, conn); errOpt != nil {
		return
	}
	opt := app.Opt
	log.Printf("handleConnectionClient: options sent: %v", opt)

	// receive ack
	//log.Printf("handleConnectionClient: FIXME WRITEME server does not send ack for UDP")
	if !app.UDP {
		var a ack
		if errAck := ackRecv(app.UDP, conn, &a); errAck != nil {
			log.Printf("handleConnectionClient: receiving ack: %v", errAck)
			return
		}
		log.Printf("handleConnectionClient: %s ack received", protoLabel(isTLS))
	}

	doneReader := make(chan struct{})
	doneWriter := make(chan struct{})

	info := ExportInfo{
		Input:  ChartData{},
		Output: ChartData{},
	}

	var input *ChartData
	var output *ChartData

	if app.CSV != "" || app.Export != "" || app.Chart != "" || app.ASCII {
		input = &info.Input
		output = &info.Output
	}

	bufSizeIn, bufSizeOut := getBufSize(opt, app.UDP)

	go clientReader(conn, c, connections, doneReader, bufSizeIn, opt, input, aggReader)
	if !app.PassiveClient {
		go clientWriter(conn, c, connections, doneWriter, bufSizeOut, opt, output, aggWriter)
	}

	tickerPeriod := time.NewTimer(app.Opt.TotalDuration)

	<-tickerPeriod.C
	log.Printf("handleConnectionClient: %v timer", app.Opt.TotalDuration)

	tickerPeriod.Stop()

	conn.Close() // force reader/writer to quit

	<-doneReader // wait reader exit
	if !app.PassiveClient {
		<-doneWriter // wait writer exit
	}

	if app.CSV != "" {

		filename := fmt.Sprintf(app.CSV, c, formatAddress(conn))
		log.Printf("exporting CSV test results to: %s", filename)
		errExport := exportCsv(filename, &info)
		if errExport != nil {
			log.Printf("handleConnectionClient: export CSV: %s: %v", filename, errExport)
		}
	}

	if app.Export != "" {
		filename := fmt.Sprintf(app.Export, c, formatAddress(conn))
		log.Printf("exporting YAML test results to: %s", filename)
		errExport := export(filename, &info)
		if errExport != nil {
			log.Printf("handleConnectionClient: export YAML: %s: %v", filename, errExport)
		}
	}

	if app.Chart != "" {
		filename := fmt.Sprintf(app.Chart, c, formatAddress(conn))
		log.Printf("rendering chart to: %s", filename)
		errRender := chartRender(filename, &info.Input, &info.Output)
		if errRender != nil {
			log.Printf("handleConnectionClient: render PNG: %s: %v", filename, errRender)
		}
	}

	plotascii(&info, conn.RemoteAddr().String(), c)

	log.Printf("handleConnectionClient: closing: %d/%d %v", c, connections, conn.RemoteAddr())
}

func getBufSize(opt Options, isUDP bool) (bufSizeIn int, bufSizeOut int) {
	if isUDP {
		bufSizeIn = opt.UDPReadSize
		bufSizeOut = opt.UDPWriteSize
		return
	}
	bufSizeIn = opt.TCPReadSize
	bufSizeOut = opt.TCPWriteSize
	return
}

func clientReader(conn net.Conn, c, connections int, done chan struct{}, bufSize int, opt Options, stat *ChartData, agg *aggregate) {
	log.Printf("clientReader: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := make([]byte, bufSize)

	workLoop(connIndex, "clientReader", "rcv/s", conn.Read, buf, opt.ReportInterval, 0, stat, agg)

	close(done)

	log.Printf("clientReader: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}

func clientWriter(conn net.Conn, c, connections int, done chan struct{}, bufSize int, opt Options, stat *ChartData, agg *aggregate) {
	log.Printf("clientWriter: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := randBuf(bufSize)

	workLoop(connIndex, "clientWriter", "snd/s", conn.Write, buf, opt.ReportInterval, opt.MaxSpeed, stat, agg)

	close(done)

	log.Printf("clientWriter: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}

func randBuf(size int) []byte {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("randBuf error: %v", err)
	}
	return buf
}

type call func(p []byte) (n int, err error)

type account struct {
	prevTime  time.Time
	prevSize  int64
	prevCalls int
	size      int64
	calls     int
}

// ChartData records data for chart
type ChartData struct {
	XValues []time.Time
	YValues []float64
}

const fmtReport = "%s %7s %14s rate: %f Mbps %6d %s"

func (a *account) update(n int, reportInterval time.Duration, conn, label, cpsLabel string, stat *ChartData, forceUpdate bool) {
	a.calls++
	a.size += int64(n)

	now := time.Now()
	elap := now.Sub(a.prevTime)
	if elap > reportInterval || forceUpdate {
		elapSec := elap.Seconds()
		mbps := float64(8*(a.size-a.prevSize)) / (1000000 * elapSec)
		cps := int64(float64(a.calls-a.prevCalls) / elapSec)
		log.Printf(fmtReport, conn, "report", label, mbps, cps, cpsLabel)
		a.prevTime = now
		a.prevSize = a.size
		a.prevCalls = a.calls

		// save chart data
		if stat != nil {
			stat.XValues = append(stat.XValues, now)
			stat.YValues = append(stat.YValues, mbps)
		}
	}
}

type aggregate struct {
	Mbps  float64 // Megabit/s
	Cps   int64   // Call/s
	Bytes int64   // total bytes
	mutex sync.Mutex
}

func (a *account) average(start time.Time, conn, label, cpsLabel string, agg *aggregate) {
	elapSec := time.Since(start).Seconds()
	mbps := float64(8*a.size) / (1000000 * elapSec)
	cps := int64(float64(a.calls) / elapSec)
	log.Printf(fmtReport, conn, "average", label, mbps, cps, cpsLabel)

	agg.mutex.Lock()
	agg.Mbps += mbps
	agg.Cps += cps
	agg.Bytes += a.size
	agg.mutex.Unlock()
}

func workLoop(conn, label, cpsLabel string, f call, buf []byte, reportInterval time.Duration, maxSpeed float64, stat *ChartData, agg *aggregate) {

	start := time.Now()
	acc := &account{}
	acc.prevTime = start

	for {
		runtime.Gosched()

		if maxSpeed > 0 {
			elapSec := time.Since(acc.prevTime).Seconds()
			if elapSec > 0 {
				mbps := float64(8*(acc.size-acc.prevSize)) / (1000000 * elapSec)
				if mbps > maxSpeed {
					time.Sleep(time.Millisecond)
					continue
				}
			}
		}

		n, errCall := f(buf)
		if errCall != nil {
			log.Printf("workLoop: %s %s: %v n=%d", conn, label, errCall, n)
			acc.update(n, reportInterval, conn, label, cpsLabel, stat, true)
			break
		}

		acc.update(n, reportInterval, conn, label, cpsLabel, stat, false)
	}

	acc.average(start, conn, label, cpsLabel, agg)
}

// Remove semi colon, invalid use in filename on windows
func formatAddress(con net.Conn) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(fmt.Sprintf("%v", con.RemoteAddr()), ":", "-", 1)
	}
	return fmt.Sprintf("%v", con.RemoteAddr())
}
