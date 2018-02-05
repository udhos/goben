package main

import (
	"log"
	"net"
	"sync"
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
			go handleConnectionClient(&wg, c, i, app.connections)
		}
	}

	wg.Wait()
}

func handleConnectionClient(wg *sync.WaitGroup, conn *net.TCPConn, c, connections int) {
	defer wg.Done()
	defer conn.Close()

	log.Printf("handleConnectionClient: starting %d/%d %v", c, connections, conn.RemoteAddr())

	log.Printf("handleConnectionClient: closing: %d/%d %v", c, connections, conn.RemoteAddr())
}
