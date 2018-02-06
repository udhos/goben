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

type message struct {
	Value int
	Bogus [10000]byte
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

type encoderWrap struct {
	writer net.Conn
	size   int64
}

func (e *encoderWrap) Write(p []byte) (n int, err error) {
	if err := e.writer.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("encoderWrap.Write: %v", err)
	}
	n, err = e.writer.Write(p)
	//log.Printf("write: %d error=%v", n, err)
	//time.Sleep(100 * time.Millisecond)
	/*
		if n < 1 {
			log.Printf("write(%d): %d error=%v", len(p), n, err)
		}
		if n < 0 {
			n = 0
		}
	*/
	e.size = int64(n)
	return
}

type decoderWrap struct {
	reader net.Conn
	size   int64
}

func (d *decoderWrap) Read(p []byte) (n int, err error) {
	if err := d.reader.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("encoderWrap.Write: %v", err)
	}
	n, err = d.reader.Read(p)
	//log.Printf("read: %d error=%v", n, err)
	/*
		if n < 1 || n > 4 {
			log.Printf("read(%d): %d error=%v", len(p), n, err)
		}
		if n < 0 {
			n = 0
		}
	*/
	d.size = int64(n)
	return
}

func serverReader(conn *net.TCPConn) {
	log.Printf("serverReader: starting: %v", conn.RemoteAddr())

	countRead := 0
	var size int64

	decoder := decoderWrap{reader: conn}
	dec := gob.NewDecoder(&decoder)
	var m message
	for {
		if err := dec.Decode(&m); err != nil {
			log.Printf("serverReader: Decode: %v", err)
			break
		}
		countRead++
		size += decoder.size
	}

	log.Printf("serverReader: exiting: %v reads=%d totalSize=%d", conn.RemoteAddr(), countRead, size)
}

func serverWriter(conn *net.TCPConn) {
	log.Printf("serverWriter: starting: %v", conn.RemoteAddr())

	countWrite := 0
	var size int64

	encoder := encoderWrap{writer: conn}
	enc := gob.NewEncoder(&encoder)
	var m message
	for {
		if err := enc.Encode(&m); err != nil {
			log.Printf("serverWriter: Encode: %v", err)
			break
		}
		countWrite++
		size += encoder.size
	}

	log.Printf("serverWriter: exiting: %v writes=%d totalSize=%d", conn.RemoteAddr(), countWrite, size)
}
