package common

func sendSerialEOFMarker(output chan<- byte, enabled bool) {
	if !enabled {
		return
	}
	for _, b := range []byte("EOF\r\n") {
		output <- b
	}
}
