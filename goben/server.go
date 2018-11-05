package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func serve(app *config) {

	var wg sync.WaitGroup

	for _, h := range app.listeners {
		hh := appendPortIfMissing(h, app.defaultPort)
		listenTCP(app, &wg, hh)
		listenUDP(app, &wg, hh)
	}

	wg.Wait()
}

func listenTCP(app *config, wg *sync.WaitGroup, h string) {
	log.Printf("serve: spawning TCP listener: %s", h)
	listener, errListen := net.Listen("tcp", h)
	if errListen != nil {
		log.Printf("listenTCP: %s: %v", h, errListen)
		return
	}
	wg.Add(1)
	go handleTCP(app, wg, listener)
}

func listenUDP(app *config, wg *sync.WaitGroup, h string) {
	log.Printf("serve: spawning UDP listener: %s", h)

	udpAddr, errAddr := net.ResolveUDPAddr("udp", h)
	if errAddr != nil {
		log.Printf("listenUDP: bad address: %s: %v", h, errAddr)
		return
	}

	conn, errListen := net.ListenUDP("udp", udpAddr)
	if errListen != nil {
		log.Printf("net.ListenUDP: %s: %v", h, errListen)
		return
	}

	wg.Add(1)
	go handleUDP(app, wg, conn)
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

func handleTCP(app *config, wg *sync.WaitGroup, listener net.Listener) {
	defer wg.Done()

	var id int

	for {
		conn, errAccept := listener.Accept()
		if errAccept != nil {
			log.Printf("handle: accept: %v", errAccept)
			break
		}
		c := conn.(*net.TCPConn)
		go handleConnection(app, c, id, 0)
		id++
	}
}

type udpInfo struct {
	remote *net.UDPAddr
	opt    options
	acc    *account
	start  time.Time
	id     int
}

func handleUDP(app *config, wg *sync.WaitGroup, conn *net.UDPConn) {
	defer wg.Done()

	tab := map[string]*udpInfo{}

	buf := make([]byte, app.opt.ReadSize)

	var idCount int

	for {
		var info *udpInfo
		n, src, errRead := conn.ReadFromUDP(buf)
		if src == nil {
			log.Printf("handleUDP: read nil src: error: %v", errRead)
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
				opt := info.opt // copy for gorouting
				go serverWriterTo(conn, opt, src, info.acc, info.id, 0)
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
			info.acc.average(info.start, connIndex, "handleUDP", "rcv/s")
			log.Printf("handleUDP: FIXME: remove idle udp entry from udp table")
			continue
		}

		// account read from UDP socket
		info.acc.update(n, info.opt.ReportInterval, connIndex, "handleUDP", "rcv/s", nil)
	}
}

func handleConnection(_ *config, conn *net.TCPConn, c, connections int) {
	defer conn.Close()

	log.Printf("handleConnection: incoming: %v", conn.RemoteAddr())

	// receive options
	var opt options
	dec := gob.NewDecoder(conn)
	if errOpt := dec.Decode(&opt); errOpt != nil {
		log.Printf("handleConnection: options failure: %v", errOpt)
		return
	}
	log.Printf("handleConnection: options received: %v", opt)

	go serverReader(conn, opt, c, connections)

	if !opt.PassiveServer {
		go serverWriter(conn, opt, c, connections)
	}

	tickerPeriod := time.NewTimer(opt.TotalDuration)

	<-tickerPeriod.C
	log.Printf("handleConnection: %v timer", opt.TotalDuration)

	tickerPeriod.Stop()

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
}

func serverReader(conn net.Conn, opt options, c, connections int) {
	log.Printf("serverReader: starting: %v", conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := make([]byte, opt.ReadSize)

	workLoop(connIndex, "serverReader", "rcv/s", conn.Read, buf, opt.ReportInterval, 0, nil)

	log.Printf("serverReader: exiting: %v", conn.RemoteAddr())
}

func serverWriter(conn net.Conn, opt options, c, connections int) {
	log.Printf("serverWriter: starting: %v", conn.RemoteAddr())

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := randBuf(opt.WriteSize)

	workLoop(connIndex, "serverWriter", "snd/s", conn.Write, buf, opt.ReportInterval, opt.MaxSpeed, nil)

	log.Printf("serverWriter: exiting: %v", conn.RemoteAddr())
}

func serverWriterTo(conn *net.UDPConn, opt options, dst net.Addr, acc *account, c, connections int) {
	log.Printf("serverWriterTo: starting: %v", dst)

	start := acc.prevTime

	udpWriteTo := func(b []byte) (int, error) {
		if time.Since(start) > opt.TotalDuration {
			return -1, fmt.Errorf("udpWriteTo: total duration %s timer", opt.TotalDuration)
		}

		return conn.WriteTo(b, dst)
	}

	connIndex := fmt.Sprintf("%d/%d", c, connections)

	buf := randBuf(opt.WriteSize)

	workLoop(connIndex, "serverWriterTo", "snd/s", udpWriteTo, buf, opt.ReportInterval, opt.MaxSpeed, nil)

	log.Printf("serverWriterTo: exiting: %v", dst)
}
