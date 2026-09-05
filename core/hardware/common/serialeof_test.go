package common

import "testing"

func TestSendSerialEOFMarker(t *testing.T) {
	output := make(chan byte, len("EOF\r\n"))

	sendSerialEOFMarker(output, true)

	if got, want := drainSerialMarker(output), "EOF\r\n"; got != want {
		t.Fatalf("marker = %q, want %q", got, want)
	}
}

func TestSendSerialEOFMarkerCanBeDisabled(t *testing.T) {
	output := make(chan byte, len("EOF\r\n"))

	sendSerialEOFMarker(output, false)

	if got := drainSerialMarker(output); got != "" {
		t.Fatalf("marker = %q, want no bytes", got)
	}
}

func drainSerialMarker(output <-chan byte) string {
	var data []byte
	for len(output) > 0 {
		data = append(data, <-output)
	}
	return string(data)
}
