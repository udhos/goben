package main

import (
	"log"
	"net"
	"sync"
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
		go handle(&wg, listener)
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

func handle(wg *sync.WaitGroup, listener net.Listener) {
	defer wg.Done()

	for {
		conn, errAccept := listener.Accept()
		if errAccept != nil {
			log.Printf("handle: accept: %v", errAccept)
			break
		}
		c := conn.(*net.TCPConn)
		go handleConnection(c)
	}
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()

	log.Printf("handleConnection: %v", conn.RemoteAddr())

	log.Printf("handleConnection: closing: %v", conn.RemoteAddr())
}
