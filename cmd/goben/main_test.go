package main

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/udhos/goben/goben"
)

func TestInvalidConfig(t *testing.T) {

	// a client config with missing TLS info, this is invalid, but validation will only happen at client startup runtime
	client := goben.NewDefaultConfig()
	client.Hosts = goben.HostList{"127.0.0.1:18443"}
	client.TLS = true
	client.TCP = false
	client.UDP = false
	client.ReportInterval = "1s"
	client.TotalDuration = "5s"
	client.Connections = 1
	client.PassiveClient = false
	assert.Equal(t, nil, goben.ValidateAndUpdateConfig(client))

	// launch client (this will fail)
	_, err := goben.Open(client)
	assert.Error(t, err)
}

func TestEndToEndInvalidTLSConfig(t *testing.T) {

	// a client config
	client := goben.NewDefaultConfig()
	client.Hosts = goben.HostList{"127.0.0.1:18444"}
	client.TLS = true
	client.TCP = false
	client.UDP = false
	client.Export = "export-%d-%s.yaml"
	client.ReportInterval = "1s"
	client.TotalDuration = "2s"
	client.Connections = 1
	client.PassiveClient = false
	client.TLSCert = "../../test/certs/client.crt"
	client.TLSKey = "../../test/certs/client.key"

	// a server config
	server := goben.NewDefaultConfig()
	server.Listeners = goben.HostList{"127.0.0.1:18444"}
	server.TLS = true
	server.TCP = false
	server.UDP = false
	server.TLSCert = "../../test/certs/ca.crt"
	server.TLSKey = "../../test/certs/ca.key"

	// launch server
	var wg sync.WaitGroup
	wg.Add(1)
	listenSuccess := goben.Serve(server, &wg)
	assert.False(t, listenSuccess)

	// launch client
	clientStats, err := goben.Open(client)
	assert.Error(t, err)
	assert.Equal(t, clientStats.TotalDuration, time.Duration(0*time.Second))
	assert.Equal(t, clientStats.ReadMbps, float64(0))
	assert.Equal(t, clientStats.WriteMbps, float64(0))
	assert.Equal(t, clientStats.ReadBytes, int64(0))
	assert.Equal(t, clientStats.WriteBytes, int64(0))
}

func TestEndToEndTLS(t *testing.T) {

	// a client config
	client := goben.NewDefaultConfig()
	client.Hosts = goben.HostList{"127.0.0.1:18445"}
	client.TLS = true
	client.TCP = false
	client.UDP = false
	client.ReportInterval = "1s"
	client.TotalDuration = "2s"
	client.Connections = 1
	client.PassiveClient = false
	client.TLSCA = "../../test/certs/ca.crt"
	client.TLSCert = "../../test/certs/client.crt"
	client.TLSKey = "../../test/certs/client.key"
	goben.ValidateAndUpdateConfig(client)
	assert.NoError(t, goben.ValidateAndUpdateConfig(client))

	// a server config
	server := goben.NewDefaultConfig()
	server.Listeners = goben.HostList{"127.0.0.1:18445"}
	server.TLS = true
	server.TCP = false
	server.UDP = false
	server.TLSCA = "../../test/certs/ca.crt"
	server.TLSCert = "../../test/certs/ca.crt"
	server.TLSKey = "../../test/certs/ca.key"
	assert.NoError(t, goben.ValidateAndUpdateConfig(server))

	// launch server
	var wg sync.WaitGroup
	wg.Add(1)
	listenSuccess := goben.Serve(server, &wg)
	if !listenSuccess {
		t.Error("server failed to listen")
	}

	// launch client
	clientStats, err := goben.Open(client)
	assert.NoError(t, err)
	assert.Equal(t, clientStats.TotalDuration, time.Duration(2*time.Second))
	assert.Greater(t, clientStats.ReadMbps, float64(100))
	assert.Greater(t, clientStats.WriteMbps, float64(100))
	assert.Greater(t, clientStats.ReadBytes, int64(100))
	assert.Greater(t, clientStats.WriteBytes, int64(100))
}

func TestEndToEndTCP(t *testing.T) {

	// a client config
	client := goben.NewDefaultConfig()
	client.Hosts = goben.HostList{"127.0.0.1:18446"}
	client.TLS = false
	client.TCP = true
	client.UDP = false
	client.ReportInterval = "1s"
	client.TotalDuration = "2s"
	client.Connections = 1
	client.PassiveClient = false

	// a server config
	server := goben.NewDefaultConfig()
	server.Listeners = goben.HostList{"127.0.0.1:18446"}
	server.TLS = false
	server.TCP = true
	server.UDP = false

	// launch server
	var wg sync.WaitGroup
	wg.Add(1)
	listenSuccess := goben.Serve(server, &wg)
	if !listenSuccess {
		t.Error("server failed to listen")
	}

	// launch client
	clientStats, err := goben.Open(client)
	assert.NoError(t, err)
	assert.Equal(t, clientStats.TotalDuration, time.Duration(2*time.Second))
	assert.Greater(t, clientStats.ReadMbps, float64(100))
	assert.Greater(t, clientStats.WriteMbps, float64(100))
	assert.Greater(t, clientStats.ReadBytes, int64(100))
	assert.Greater(t, clientStats.WriteBytes, int64(100))
}

func TestEndToEndTCPFallback(t *testing.T) {

	// a client config
	client := goben.NewDefaultConfig()
	client.Hosts = goben.HostList{"127.0.0.1:18447"}
	client.TLS = true
	client.TCP = true
	client.UDP = false
	client.ReportInterval = "1s"
	client.TotalDuration = "2s"
	client.Connections = 1
	client.PassiveClient = false

	// a server config
	server := goben.NewDefaultConfig()
	server.Listeners = goben.HostList{"127.0.0.1:18447"}
	server.TLS = true
	server.TCP = true
	server.UDP = false

	// launch server
	var wg sync.WaitGroup
	wg.Add(1)
	listenSuccess := goben.Serve(server, &wg)
	if !listenSuccess {
		t.Error("server failed to listen")
	}

	// launch client
	clientStats, err := goben.Open(client)
	assert.NoError(t, err)
	assert.Equal(t, clientStats.TotalDuration, time.Duration(2*time.Second))
	assert.Greater(t, clientStats.ReadMbps, float64(100))
	assert.Greater(t, clientStats.WriteMbps, float64(100))
	assert.Greater(t, clientStats.ReadBytes, int64(100))
	assert.Greater(t, clientStats.WriteBytes, int64(100))
}
