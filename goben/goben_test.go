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

func TestDefaultTimeUnit(t *testing.T) {
	if defaultTimeUnit("") != "" {
		t.Errorf("empty time unit should be preserved")
	}
	if defaultTimeUnit(" ") != " " {
		t.Errorf("blank time unit should be preserved")
	}
	if defaultTimeUnit("10s") != "10s" {
		t.Errorf("explicit seconds time unit should be preserved")
	}
	if defaultTimeUnit("10m") != "10m" {
		t.Errorf("explicit minutes time unit should be preserved")
	}
	if defaultTimeUnit("10") != "10s" {
		t.Errorf("implicit time unit should default to seconds")
	}
}
