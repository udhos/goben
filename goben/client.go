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

	doneReader := make(chan struct{})
	doneWriter := make(chan struct{})

	go clientReader(conn, c, connections, doneReader)
	go clientWriter(conn, c, connections, doneWriter)

	tickerReport := time.NewTicker(app.durationReportInterval)
	tickerPeriod := time.NewTimer(app.durationTotalDuration)

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

	tickerReport.Stop()
	tickerPeriod.Stop()

	conn.Close() // force reader/writer to quit

	<-doneReader // wait reader exit
	<-doneWriter // wait writer exit

	log.Printf("handleConnectionClient: closing: %d/%d %v", c, connections, conn.RemoteAddr())
}

func clientReader(conn *net.TCPConn, c, connections int, done chan struct{}) {
	log.Printf("clientReader: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	countRead := 0

	dec := gob.NewDecoder(conn)
	var m message
	for {
		if err := dec.Decode(&m); err != nil {
			log.Printf("clientReader: Decode: %v", err)
			break
		}
		countRead++
	}

	close(done)

	log.Printf("clientReader: exiting: %d/%d %v reads=%d", c, connections, conn.RemoteAddr(), countRead)
}

func clientWriter(conn *net.TCPConn, c, connections int, done chan struct{}) {
	log.Printf("clientWriter: starting: %d/%d %v", c, connections, conn.RemoteAddr())

	countWrite := 0

	enc := gob.NewEncoder(conn)
	var m message
	for {
		if err := enc.Encode(&m); err != nil {
			log.Printf("clientWriter: Encode: %v", err)
			break
		}
		countWrite++
	}

	close(done)

	log.Printf("clientWriter: exiting: %d/%d %v writes=%d", c, connections, conn.RemoteAddr(), countWrite)

}
