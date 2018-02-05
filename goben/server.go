package main

import (
	"log"
	"net"
	"sync"
)

func serve(app *config) {

	var wg sync.WaitGroup

	for _, h := range app.listeners {

		log.Printf("serve: spawning TCP listener: %s", h)

		listener, errListen := net.Listen("tcp", h)
		if errListen != nil {
			log.Printf("serve: listen: %v", errListen)
			continue
		}

		wg.Add(1)
		go handle(&wg, listener)
	}

	wg.Wait()
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
}
