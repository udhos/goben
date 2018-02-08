package main

import (
	"encoding/gob"
	//"io"
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

func handleConnection(app *config, conn *net.TCPConn) {
	defer conn.Close()

	log.Printf("handleConnection: incoming: %v", conn.RemoteAddr())

	// receive options
	var opt options
	dec := gob.NewDecoder(conn)
	if errOpt := dec.Decode(&opt); errOpt != nil {
		log.Printf("handleConnection: options failure: %v", errOpt)
		return
	}
	log.Printf("handleConnection: options: %v", opt)

	go serverReader(conn, opt)

	if !opt.PassiveServer {
		go serverWriter(conn, opt)
	}

	tickerPeriod := time.NewTimer(opt.TotalDuration)

	<-tickerPeriod.C
	log.Printf("handleConnection: %v timer", opt.TotalDuration)

	tickerPeriod.Stop()

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
}

func serverReader(conn *net.TCPConn, opt options) {
	log.Printf("serverReader: starting: %v", conn.RemoteAddr())

	workLoop("serverReader", "rcv/s", conn.Read, opt.ReadSize, opt.ReportInterval, 0)

	log.Printf("serverReader: exiting: %v", conn.RemoteAddr())
}

func serverWriter(conn *net.TCPConn, opt options) {
	log.Printf("serverWriter: starting: %v", conn.RemoteAddr())

	workLoop("serverWriter", "snd/s", conn.Write, opt.WriteSize, opt.ReportInterval, opt.MaxSpeed)

	log.Printf("serverWriter: exiting: %v", conn.RemoteAddr())
}
