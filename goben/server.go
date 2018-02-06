package main

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
	"time"
)

func serve(app *config) {

	var wg sync.WaitGroup

	for _, h := range app.listeners {

		hh := appendPortIfMissing(h, app.defaultPort)

		log.Printf("serve: spawning TCP listener: %s", hh)

		listener, errListen := net.Listen("tcp", hh)
		if errListen != nil {
			log.Printf("serve: listen: %s: %v", hh, errListen)
			continue
		}

		wg.Add(1)
		go handle(app, &wg, listener)
	}

	wg.Wait()
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

func handle(app *config, wg *sync.WaitGroup, listener net.Listener) {
	defer wg.Done()

	for {
		conn, errAccept := listener.Accept()
		if errAccept != nil {
			log.Printf("handle: accept: %v", errAccept)
			break
		}
		c := conn.(*net.TCPConn)
		go handleConnection(app, c)
	}
}

type message struct {
	Value int
}

func handleConnection(app *config, conn *net.TCPConn) {
	defer conn.Close()

	log.Printf("handleConnection: incoming: %v", conn.RemoteAddr())

	go serverReader(conn)
	go serverWriter(conn)

	tickerReport := time.NewTicker(app.durationReportInterval)
	tickerPeriod := time.NewTimer(app.durationTotalDuration)

	// timer loop
LOOP:
	for {
		select {
		case <-tickerReport.C:
			log.Printf("handleConnection: tick")
		case <-tickerPeriod.C:
			log.Printf("handleConnection: timer")
			break LOOP
		}
	}

	tickerReport.Stop()
	tickerPeriod.Stop()

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
}

func serverReader(conn *net.TCPConn) {
	log.Printf("serverReader: starting: %v", conn.RemoteAddr())

	countRead := 0

	dec := gob.NewDecoder(conn)
	var m message
	for {
		if err := dec.Decode(&m); err != nil {
			log.Printf("serverReader: Decode: %v", err)
			break
		}
		countRead++
	}

	log.Printf("serverReader: exiting: %v reads=%d", conn.RemoteAddr(), countRead)
}

func serverWriter(conn *net.TCPConn) {
	log.Printf("serverWriter: starting: %v", conn.RemoteAddr())

	countWrite := 0

	enc := gob.NewEncoder(conn)
	var m message
	for {
		if err := enc.Encode(&m); err != nil {
			log.Printf("serverWriter: Encode: %v", err)
			break
		}
		countWrite++
	}

	log.Printf("serverWriter: exiting: %v writes=%d", conn.RemoteAddr(), countWrite)

}
