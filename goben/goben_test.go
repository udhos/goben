package main

import (
	"testing"
)

func TestAppendPort(t *testing.T) {
	expectAppendPort(t, "", "", "")
	expectAppendPort(t, "", ":80", ":80")
	//expectAppendPort(t, ":", ":80", ":80")

	expectAppendPort(t, "localhost", ":80", "localhost:80")
	expectAppendPort(t, "localhost:8080", ":80", "localhost:8080")
	//expectAppendPort(t, "localhost:", ":80", "localhost:80")

	expectAppendPort(t, "127.0.0.1", ":80", "127.0.0.1:80")
	expectAppendPort(t, "127.0.0.1:8080", ":80", "127.0.0.1:8080")
	//expectAppendPort(t, "127.0.0.1:", ":80", "127.0.0.1:80")

	expectAppendPort(t, "[::1]", ":80", "[::1]:80")
	expectAppendPort(t, "[::1]:8080", ":80", "[::1]:8080")
	//expectAppendPort(t, "[::1]:", ":80", "[::1]:80")
}

func expectAppendPort(t *testing.T, host, port, wanted string) {
	result := appendPortIfMissing(host, port)
	if result != wanted {
		t.Errorf("expectAppendPort: host=%s port=%s result=%s wanted=%s",
			host, port, result, wanted)
	}
}
