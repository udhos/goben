package main

import (
	"encoding/gob"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

func open(app *config) {

	var wg sync.WaitGroup

	for _, h := range app.hosts {

		hh := appendPortIfMissing(h, app.defaultPort)

		for i := 0; i < app.connections; i++ {

			var proto string
			if app.udp {
				proto = "udp"
			} else {
				proto = "tcp"
			}

			log.Printf("open: opening %s %d/%d: %s", proto, i, app.connections, hh)

			conn, errDial := net.Dial(proto, hh)
			if errDial != nil {
				log.Printf("open: dial %s: %s: %v", proto, hh, errDial)
				continue
			}

			wg.Add(1)
			//c := conn.(*net.TCPConn)
			go handleConnectionClient(app, &wg, conn, i, app.connections)
		}
	}

	wg.Wait()
}

func handleConnectionClient(app *config, wg *sync.WaitGroup, conn net.Conn, c, connections int) {
	defer wg.Done()

	log.Printf("handleConnectionClient: starting %d/%d %v", c, connections, conn.RemoteAddr())

	// send options
	opt := app.opt
	enc := gob.NewEncoder(conn)
	if errOpt := enc.Encode(&opt); errOpt != nil {
		log.Printf("handleConnectionClient: options failure: %v", errOpt)
		return
	}
	log.Printf("handleConnectionClient: options sent: %v", opt)

	doneReader := make(chan struct{})
	doneWriter := make(chan struct{})

	go clientReader(conn, c, connections, doneReader, opt)
	if !app.passiveClient {
		go clientWriter(conn, c, connections, doneWriter, opt)
	}

	tickerPeriod := time.NewTimer(app.opt.TotalDuration)

	<-tickerPeriod.C
	log.Printf("handleConnectionClient: %v timer", app.opt.TotalDuration)

	tickerPeriod.Stop()

	conn.Close() // force reader/writer to quit

	<-doneReader // wait reader exit
	if !app.passiveClient {
		<-doneWriter // wait writer exit
	}

	log.Printf("handleConnectionClient: closing: %d/%d %v", c, connections, conn.RemoteAddr())
}

func clientReader(conn net.Conn, c, connections int, done chan struct{}, opt options) {
	log.Printf("clientReader: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	workLoop("clientReader", "rcv/s", conn.Read, opt.ReadSize, opt.ReportInterval, 0)

	close(done)

	log.Printf("clientReader: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}

func clientWriter(conn net.Conn, c, connections int, done chan struct{}, opt options) {
	log.Printf("clientWriter: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	workLoop("clientWriter", "snd/s", conn.Write, opt.WriteSize, opt.ReportInterval, opt.MaxSpeed)

	close(done)

	log.Printf("clientWriter: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}

type call func(p []byte) (n int, err error)

func workLoop(label, cpsLabel string, f call, bufSize int, reportInterval time.Duration, maxSpeed float64) {

	countCalls := 0
	var size int64

	buf := make([]byte, bufSize)

	start := time.Now()
	prevTime := start
	prevSize := size
	prevCount := countCalls

	for {
		runtime.Gosched()

		if maxSpeed > 0 {
			elapSec := time.Since(prevTime).Seconds()
			if elapSec > 0 {
				mbps := float64(8*(size-prevSize)) / (1000000 * elapSec)
				if mbps > maxSpeed {
					time.Sleep(time.Millisecond)
					continue
				}
			}
		}

		n, errCall := f(buf)
		if errCall != nil {
			log.Printf("workLoop: %s: %v", label, errCall)
			break
		}
		countCalls++
		size += int64(n)

		now := time.Now()
		elap := now.Sub(prevTime)
		if elap > reportInterval {
			elapSec := elap.Seconds()
			mbps := int64(float64(8*(size-prevSize)) / (1000000 * elapSec))
			cps := int64(float64(countCalls-prevCount) / elapSec)
			log.Printf("report %s rate: %6d Mbps %6d %s", label, mbps, cps, cpsLabel)
			prevTime = now
			prevSize = size
			prevCount = countCalls
		}
	}

	elapSec := time.Since(start).Seconds()
	mbps := int64(float64(8*size) / (1000000 * elapSec))
	cps := int64(float64(countCalls) / elapSec)
	log.Printf("average %s rate: %d Mbps %d %s", label, mbps, cps, cpsLabel)
}
