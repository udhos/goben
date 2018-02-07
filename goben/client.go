package main

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
	"time"
)

func open(app *config) {

	var wg sync.WaitGroup

	for _, h := range app.hosts {

		hh := appendPortIfMissing(h, app.defaultPort)

		for i := 0; i < app.connections; i++ {

			log.Printf("open: opening TCP %d/%d: %s", i, app.connections, hh)

			conn, errDial := net.Dial("tcp", hh)
			if errDial != nil {
				log.Printf("open: dial: %s: %v", hh, errDial)
				continue
			}

			wg.Add(1)
			c := conn.(*net.TCPConn)
			go handleConnectionClient(app, &wg, c, i, app.connections)
		}
	}

	wg.Wait()
}

func handleConnectionClient(app *config, wg *sync.WaitGroup, conn *net.TCPConn, c, connections int) {
	defer wg.Done()
	//defer conn.Close()

	log.Printf("handleConnectionClient: starting %d/%d %v", c, connections, conn.RemoteAddr())

	// send options
	opt := app.opt
	enc := gob.NewEncoder(conn)
	if errOpt := enc.Encode(&opt); errOpt != nil {
		log.Printf("handleConnectionClient: options failure: %v", errOpt)
		return
	}

	doneReader := make(chan struct{})
	doneWriter := make(chan struct{})

	go clientReader(conn, c, connections, doneReader, opt)
	if !app.passiveClient {
		go clientWriter(conn, c, connections, doneWriter, opt)
	}

	//tickerReport := time.NewTicker(app.opt.ReportInterval)
	tickerPeriod := time.NewTimer(app.opt.TotalDuration)

	/*
			// timer loop
		LOOP:
				for {
					select {
					case <-tickerReport.C:
						log.Printf("handleConnectionClient: tick")
					case <-tickerPeriod.C:
						log.Printf("handleConnectionClient: timer")
						break LOOP
					}
				}
	*/

	<-tickerPeriod.C
	log.Printf("handleConnectionClient: %v timer", app.opt.TotalDuration)

	//tickerReport.Stop()
	tickerPeriod.Stop()

	conn.Close() // force reader/writer to quit

	<-doneReader // wait reader exit
	if !app.passiveClient {
		<-doneWriter // wait writer exit
	}

	log.Printf("handleConnectionClient: closing: %d/%d %v", c, connections, conn.RemoteAddr())
}

func clientReader(conn *net.TCPConn, c, connections int, done chan struct{}, opt options) {
	log.Printf("clientReader: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	/*
		countCalls := 0
		var size int64

		buf := make([]byte, opt.ReadSize)

		prevTime := time.Now()
		prevSize := size
		prevCount := countCalls

		for {
			n, errRead := conn.Read(buf)
			if errRead != nil {
				log.Printf("clientReader: Read: %v", errRead)
				break
			}
			countCalls++
			size += int64(n)

			now := time.Now()
			elap := now.Sub(prevTime)
			if elap > opt.ReportInterval {
				mbps := int64(8 * float64(size-prevSize) / (1000000 * elap.Seconds()))
				log.Printf("report calls=%d input=%v Mbps", countCalls-prevCount, mbps)
				prevTime = now
				prevSize = size
				prevCount = countCalls
			}
		}
	*/

	workLoop("clientReader", conn.Read, opt.ReadSize, opt.ReportInterval)

	close(done)

	log.Printf("clientReader: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}

type call func(p []byte) (n int, err error)

func workLoop(label string, f call, bufSize int, reportInterval time.Duration) {

	countCalls := 0
	var size int64

	buf := make([]byte, bufSize)

	prevTime := time.Now()
	prevSize := size
	prevCount := countCalls

	for {
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
			mbps := int64(8 * float64(size-prevSize) / (1000000 * elap.Seconds()))
			log.Printf("report %s calls=%d rate=%v Mbps", label, countCalls-prevCount, mbps)
			prevTime = now
			prevSize = size
			prevCount = countCalls
		}
	}
}

func clientWriter(conn *net.TCPConn, c, connections int, done chan struct{}, opt options) {
	log.Printf("clientWriter: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	/*
		countCalls := 0
		var size int64

		prevTime := time.Now()
		prevSize := size
		prevCount := countCalls

		buf := make([]byte, opt.WriteSize)
		for {
			n, errWrite := conn.Write(buf)
			if errWrite != nil {
				log.Printf("clientWriter: Write: %v", errWrite)
				break
			}
			countCalls++
			size += int64(n)

			now := time.Now()
			elap := now.Sub(prevTime)
			if elap > opt.ReportInterval {
				mbps := int64(8 * float64(size-prevSize) / (1000000 * elap.Seconds()))
				log.Printf("report calls=%d output=%v Mbps", countCalls-prevCount, mbps)
				prevTime = now
				prevSize = size
				prevCount = countCalls
			}
		}
	*/

	workLoop("clientWriter", conn.Write, opt.WriteSize, opt.ReportInterval)

	close(done)

	log.Printf("clientWriter: exiting: %d/%d %v", c, connections, conn.RemoteAddr())
}
