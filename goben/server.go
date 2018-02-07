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

/*
type message struct {
	Value int
	Bogus [10000]byte
}
*/

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

	//tickerReport := time.NewTicker(opt.ReportInterval)
	tickerPeriod := time.NewTimer(opt.TotalDuration)

	/*
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
	*/

	<-tickerPeriod.C
	log.Printf("handleConnection: %v timer", opt.TotalDuration)

	//tickerReport.Stop()
	tickerPeriod.Stop()

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
}

/*
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
	d.size = int64(n)
	return
}
*/

func serverReader(conn *net.TCPConn, opt options) {
	log.Printf("serverReader: starting: %v", conn.RemoteAddr())

	/*
		countRead := 0
		var size int64

		buf := make([]byte, opt.ReadSize)
		for {
			n, errRead := conn.Read(buf)
			if errRead != nil {
				log.Printf("serverReader: Read: %v", errRead)
				break
			}
			countRead++
			size += int64(n)
		}
	*/

	workLoop("serverReader", conn.Read, opt.ReadSize, opt.ReportInterval)

	log.Printf("serverReader: exiting: %v", conn.RemoteAddr())
}

func serverWriter(conn *net.TCPConn, opt options) {
	log.Printf("serverWriter: starting: %v", conn.RemoteAddr())

	/*
		countWrite := 0
		var size int64

		buf := make([]byte, opt.WriteSize)
		for {
			n, errWrite := conn.Write(buf)
			if errWrite != nil {
				log.Printf("serverWriter: Write: %v", errWrite)
				break
			}
			countWrite++
			size += int64(n)
		}
	*/

	workLoop("serverWriter", conn.Write, opt.WriteSize, opt.ReportInterval)

	log.Printf("serverWriter: exiting: %v", conn.RemoteAddr())
}
