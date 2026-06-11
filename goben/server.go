package goben

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// Serve starts a server with a given configuration.
func Serve(ctx context.Context, app *Config, wg *sync.WaitGroup) bool {

	// validate the config
	if err := ValidateAndUpdateConfig(app); err != nil {
		return false
	}

	// support falling back to TCP mode
	if app.TLS && !fileExists(app.TLSKey) {
		log.Printf("key file not found: %s - disabling TLS", app.TLSKey)
		app.TLS = false
	}
	if app.TLS && !fileExists(app.TLSCert) {
		log.Printf("cert file not found: %s - disabling TLS", app.TLSCert)
		app.TLS = false
	}
	if app.TLS && !fileExists(app.TLSCA) {
		log.Printf("CA file not found: %s - disabling TLS", app.TLSCA)
		app.TLS = false
	}

	// check if any mode remains
	if !app.TLS && !app.TCP && !app.UDP {
		return false
	}

	successfulListeners := 0
	for _, h := range app.Listeners {
		hh := appendPortIfMissing(h, app.DefaultPort)
		tcpSuccess := listenTCP(ctx, app, wg, hh)
		udpSuccess := listenUDP(ctx, app, wg, hh)
		if tcpSuccess || udpSuccess {
			successfulListeners++
		}
	}

	// if no listeners were successful, return that the socket is not listening
	if successfulListeners == 0 {
		log.Print("serve: no listeners successful")
		return false
	}

	// return that the socket is now listening
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func listenTCP(ctx context.Context, app *Config, wg *sync.WaitGroup, h string) bool {

	// first try TLS
	if app.TLS {
		log.Printf("listenTCP: spawning TLS listener: %s", h)
		listener, errTLS := listenTLS(app, h)
		if errTLS == nil {
			spawnAcceptLoopTCP(ctx, wg, listener, true)
			return true
		}
		log.Printf("listenTLS: %v", errTLS)
		// TLS failed, try plain TCP if enabled
		if !app.TCP {
			log.Print("listenTCP: TLS failed and TCP disabled")
			return false
		}
	} else {
		log.Print("listenTCP: TLS disabled")
	}

	// only use TCP if explicitly enabled
	if app.TCP {
		if app.TLS {
			log.Printf("listenTCP: falling back to TCP and spawning TCP listener: %s", h)
		} else {
			log.Printf("listenTCP: spawning TLS listener: %s", h)
		}
		listener, errListen := net.Listen("tcp", h)
		if errListen != nil {
			log.Printf("listenTCP: TLS=%v %s: %v", app.TLS, h, errListen)
			return false
		}
		spawnAcceptLoopTCP(ctx, wg, listener, false)
		return true
	}

	log.Print("listenTCP: TCP disabled")

	return false
}

func spawnAcceptLoopTCP(ctx context.Context, wg *sync.WaitGroup, listener net.Listener, isTLS bool) {
	wg.Add(1)
	go handleTCP(ctx, wg, listener, isTLS)
}

func listenTLS(app *Config, h string) (net.Listener, error) {
	log.Printf("reading cert and key from %s %s", app.TLSCert, app.TLSKey)

	// load the server cert
	cert, errCert := tls.LoadX509KeyPair(app.TLSCert, app.TLSKey)
	if errCert != nil {
		log.Printf("listenTLS: failure loading TLS key pair: %v", errCert)
		app.TLS = false // disable TLS
		return nil, errCert
	}

	// load client CA cert
	caCert, err := os.ReadFile(app.TLSCA)
	if err != nil {
		log.Printf("listenTLS: failure reading CA cert: %v", err)
		return nil, err
	}

	// create a cert pool
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// set the client auth settings
	clientAuth := tls.RequireAndVerifyClientCert
	if !app.TLSAuthClient {
		clientAuth = tls.NoClientCert
	}

	// create the TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   clientAuth,
	}
	listener, errListen := tls.Listen("tcp", h, config)
	return listener, errListen
}

func listenUDP(ctx context.Context, app *Config, wg *sync.WaitGroup, h string) bool {
	if app.UDP {
		log.Printf("serve: spawning UDP listener: %s", h)

		udpAddr, errAddr := net.ResolveUDPAddr("udp", h)
		if errAddr != nil {
			log.Printf("listenUDP: bad address: %s: %v", h, errAddr)
			return false
		}

		conn, errListen := net.ListenUDP("udp", udpAddr)
		if errListen != nil {
			log.Printf("net.ListenUDP: %s: %v", h, errListen)
			return false
		}

		wg.Add(1)
		go handleUDP(ctx, app, wg, conn)
	} else {
		log.Print("listenUDP: UDP disabled")
		return false
	}
	return true
}

func appendPortIfMissing(host, port string) string {

LOOP:
	for i := len(host) - 1; i >= 0; i-- {
		c := host[i]
		switch c {
		case ']':
			break LOOP
		case ':':
			/*
				if i == len(host)-1 {
					return host[:len(host)-1] + port // drop repeated :
				}
			*/
			return host
		}
	}

	return host + port
}

func handleTCP(ctx context.Context, wg *sync.WaitGroup, listener net.Listener, isTLS bool) {
	defer wg.Done()

	// Use a derived context so the closer goroutine exits when handleTCP returns,
	// even if the parent context has not been cancelled yet.
	ctx, stopCloser := context.WithCancel(ctx)
	defer stopCloser()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	var id int

	var aggReader aggregate
	var aggWriter aggregate

	var retryDelay time.Duration

	for {
		conn, errAccept := listener.Accept()
		if errAccept != nil {
			select {
			case <-ctx.Done():
				log.Printf("handleTCP: shutdown requested")
				return
			default:
			}
			if errors.Is(errAccept, net.ErrClosed) {
				return
			}
			// Retry transient errors with exponential backoff.
			if retryDelay == 0 {
				retryDelay = 5 * time.Millisecond
			} else {
				retryDelay *= 2
			}
			if retryDelay > time.Second {
				log.Printf("handle: accept: %v", errAccept)
				return
			}
			log.Printf("handle: accept temporary error, retrying in %v: %v", retryDelay, errAccept)
			time.Sleep(retryDelay)
			continue
		}
		retryDelay = 0
		go handleConnection(ctx, conn, id, 0, isTLS, &aggReader, &aggWriter)
		id++
	}
}

type udpInfo struct {
	remote *net.UDPAddr
	opt    Options
	acc    *account
	start  time.Time
	id     int
}

func handleUDP(ctx context.Context, app *Config, wg *sync.WaitGroup, conn *net.UDPConn) {
	defer wg.Done()

	// Use a derived context so the closer goroutine exits when handleUDP returns,
	// even if the parent context has not been cancelled yet.
	ctx, stopCloser := context.WithCancel(ctx)
	defer stopCloser()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	tab := map[string]*udpInfo{}

	buf := make([]byte, app.Opt.UDPReadSize)

	var aggReader aggregate
	var aggWriter aggregate

	var idCount int

	for {
		var info *udpInfo
		n, src, errRead := conn.ReadFromUDP(buf)
		if errRead != nil {
			select {
			case <-ctx.Done():
				log.Printf("handleUDP: shutdown requested")
				return
			default:
				if src == nil {
					log.Printf("handleUDP: read nil src: error: %v", errRead)
					continue
				}
				log.Printf("handleUDP: read error: %v", errRead)
				continue
			}
		}
		if src == nil {
			continue
		}
		var found bool
		info, found = tab[src.String()]
		if !found {
			log.Printf("handleUDP: incoming: %v", src)

			info = &udpInfo{
				remote: src,
				acc:    &account{},
				start:  time.Now(),
				id:     idCount,
			}
			idCount++
			info.acc.prevTime = info.start
			tab[src.String()] = info

			dec := gob.NewDecoder(bytes.NewBuffer(buf[:n]))
			if errOpt := dec.Decode(&info.opt); errOpt != nil {
				log.Printf("handleUDP: options failure: %v", errOpt)
				continue
			}
			log.Printf("handleUDP: options received: %v", info.opt)

			if !info.opt.PassiveServer {
				opt := info.opt // copy for goroutine
				go serverWriterTo(ctx, conn, opt, src, info.acc, info.id, 0, &aggWriter)
			}

			continue
		}

		connIndex := fmt.Sprintf("%d/%d", info.id, 0)

		if errRead != nil {
			log.Printf("handleUDP: %s read error: %s: %v", connIndex, src, errRead)
			continue
		}

		if time.Since(info.start) > info.opt.TotalDuration {
			log.Printf("handleUDP: total duration %s timer: %s", info.opt.TotalDuration, src)
			info.acc.average(info.start, connIndex, "handleUDP", "rcv/s", &aggReader)
			log.Printf("handleUDP: FIXME: remove idle udp entry from udp table")
			continue
		}

		info.acc.update(n, info.opt.ReportInterval, connIndex, "handleUDP", "rcv/s", nil, false)
	}
}

func handleConnection(ctx context.Context, conn net.Conn, c, connections int, isTLS bool, aggReader, aggWriter *aggregate) {
	// Use sync.Once so conn.Close() is safe to call explicitly before returning
	// (to unblock goroutines) as well as via defer for early-exit paths.
	var closeOnce sync.Once
	closeConn := func() { closeOnce.Do(func() { conn.Close() }) }
	defer closeConn()

	log.Printf("handleConnection: incoming: %s %v", protoLabel(isTLS), conn.RemoteAddr())

	// ensure the TLS handshake if this is a TLS connection
	tlscon, ok := conn.(*tls.Conn)
	if ok {
		err := tlscon.Handshake()
		if err != nil {
			log.Printf("server: handshake failed: %v", err)
			return
		}
		state := tlscon.ConnectionState()
		log.Println("Server: client public key is:")
		for _, v := range state.PeerCertificates {
			log.Print("- Subject: ", v.Subject)
			log.Print("  Issuer: ", v.Issuer)
			log.Print("  Expiration: ", v.NotAfter)
		}
	} else {
		log.Print("handleConnection: not TLS")
	}

	// receive options
	var opt Options
	dec := gob.NewDecoder(conn)
	if errOpt := dec.Decode(&opt); errOpt != nil {
		if isTLS {
			log.Printf("handleConnection: options failure: %v", errOpt)
		} else {
			log.Printf("handleConnection: options failure - it might be client attempting our (disabled) TLS first: %v", errOpt)
		}
		return
	}
	log.Printf("handleConnection: options received: %v", opt)

	if clientVersion, ok := opt.Table["clientVersion"]; ok {
		log.Printf("handleConnection: clientVersion=%s", clientVersion)
	}

	// send ack
	a := newAck()
	if errAck := ackSend(false, conn, a); errAck != nil {
		log.Printf("handleConnection: sending ack: %v", errAck)
		return
	}

	var connWg sync.WaitGroup

	connWg.Go(func() {
		serverReader(ctx, conn, opt, c, connections, isTLS, aggReader)
	})

	if !opt.PassiveServer {
		connWg.Go(func() {
			serverWriter(ctx, conn, opt, c, connections, isTLS, aggWriter)
		})
	}

	tickerPeriod := time.NewTimer(opt.TotalDuration)

	select {
	case <-tickerPeriod.C:
		log.Printf("handleConnection: %v timer", opt.TotalDuration)
	case <-ctx.Done():
		log.Printf("handleConnection: received shutdown signal")
	}

	tickerPeriod.Stop()

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
	closeConn() // force reader/writer goroutines to unblock
	connWg.Wait()
}

func serverReader(ctx context.Context, conn net.Conn, opt Options, c, connections int, isTLS bool, agg *aggregate) {

	log.Printf("serverReader: starting: %s %v", protoLabel(isTLS), conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := make([]byte, opt.TCPReadSize)

	workLoop(ctx, connIndex, "serverReader", "rcv/s", conn.Read, buf, opt.ReportInterval, 0, nil, agg)

	log.Printf("serverReader: exiting: %v", conn.RemoteAddr())
}

func protoLabel(isTLS bool) string {
	if isTLS {
		return "TLS"
	}
	return "TCP"
}

func serverWriter(ctx context.Context, conn net.Conn, opt Options, c, connections int, isTLS bool, agg *aggregate) {

	log.Printf("serverWriter: starting: %s %v", protoLabel(isTLS), conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := randBuf(opt.TCPWriteSize)

	workLoop(ctx, connIndex, "serverWriter", "snd/s", conn.Write, buf, opt.ReportInterval, opt.MaxSpeed, nil, agg)

	log.Printf("serverWriter: exiting: %v", conn.RemoteAddr())
}

func serverWriterTo(ctx context.Context, conn *net.UDPConn, opt Options, dst net.Addr, acc *account, c, connections int, agg *aggregate) {
	log.Printf("serverWriterTo: starting: UDP %v", dst)

	start := acc.prevTime

	udpWriteTo := func(b []byte) (int, error) {
		if time.Since(start) > opt.TotalDuration {
			return -1, fmt.Errorf("udpWriteTo: total duration %s timer", opt.TotalDuration)
		}

		return conn.WriteTo(b, dst)
	}

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := randBuf(opt.UDPWriteSize)

	workLoop(ctx, connIndex, "serverWriterTo", "snd/s", udpWriteTo, buf, opt.ReportInterval, opt.MaxSpeed, nil, agg)

	log.Printf("serverWriterTo: exiting: %v", dst)
}
