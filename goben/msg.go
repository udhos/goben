package goben

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"time"
)

// Options are sent to configure server behavior.
type Options struct {
	ReportInterval time.Duration
	TotalDuration  time.Duration
	TCPReadSize    int
	TCPWriteSize   int
	UDPReadSize    int
	UDPWriteSize   int
	PassiveServer  bool              // suppress server send
	MaxSpeed       float64           // mbps
	Table          map[string]string // send optional information client->server
}

type ack struct {
	Magic string
	Table map[string]string // send optional information server->client
}

const ackMagic = "goben-ack"

func newAck() ack {
	a := ack{
		Magic: ackMagic,
		Table: map[string]string{
			"serverVersion": Version,
		},
	}
	return a
}

// ackSend server sends
func ackSend(udp bool, conn io.Writer, a ack) error {

	// prevent sending wrong magic
	if a.Magic != ackMagic {
		m := fmt.Sprintf("ackSend: bad magic: expected=[%s] got=[%s]", ackMagic, a.Magic)
		log.Print(m)
		return fmt.Errorf(m)
	}

	if udp {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if errEnc := enc.Encode(&a); errEnc != nil {
			log.Printf("ackSend: UDP encoding: %v", errEnc)
			return errEnc
		}
		_, errWrite := conn.Write(buf.Bytes())
		if errWrite != nil {
			log.Printf("ackSend: UDP write: %v", errWrite)
			return errWrite
		}
		return nil
	}

	enc := gob.NewEncoder(conn)
	if errEnc := enc.Encode(&a); errEnc != nil {
		log.Printf("ackSend: TCP failure: %v", errEnc)
		return errEnc
	}

	return nil
}

// ackRecv client receives
func ackRecv(udp bool, conn io.Reader, a *ack) error {

	if udp {
		const m = "ackRecv: UDP FIXME WRITEME"
		log.Print(m)
		return fmt.Errorf(m)
	}

	dec := gob.NewDecoder(conn)
	if errDec := dec.Decode(a); errDec != nil {
		log.Printf("ackRecv: TCP failure: %v", errDec)
		return errDec
	}

	// prevent receiving wrong magic
	if a.Magic != ackMagic {
		m := fmt.Sprintf("ackRecv: bad magic: expected=[%s] got=[%s]", ackMagic, a.Magic)
		log.Print(m)
		return fmt.Errorf(m)
	}

	if serverVersion, ok := a.Table["serverVersion"]; ok {
		log.Printf("serverVersion=%s", serverVersion)
	}

	return nil
}
