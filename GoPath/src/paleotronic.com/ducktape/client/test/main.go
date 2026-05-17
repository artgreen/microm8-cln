package main

import (
	"time"

	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
)

func main() {
	c := client.NewDuckTapeClient("localhost", ":9988", "narf", "udp")

	err := c.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %s\n", err.Error())
	}

	log.Println("Connected.")

	time.Sleep(time.Millisecond * 100)

	defer c.Close()

	c.SendMessage("SUB", []byte("chicken"), false)
	c.SendMessage("SND", []byte("chicken-control"), false)

	for c.Connected {
		select {
		case msg := <-c.Incoming:
			log.Println(msg)
		}
	}
}
