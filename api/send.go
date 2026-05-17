package s8webclient

import (
	"time"

	"paleotronic.com/ducktape"
)

// SendAndWait sends a network request with a timeout
func (c *Client) SendAndWait(id string, payload []byte, valid []string) (*ducktape.DuckTapeBundle, error) {

	//	var err error

	if c.c == nil {
		return &ducktape.DuckTapeBundle{}, ErrNotConnected
	}

	// Now do the connection
	c.c.SendMessage(id, payload, true)

	// get response
	tochan := time.After(time.Second * 20)
	//var bb []byte
	select {
	case _ = <-tochan:
		return &ducktape.DuckTapeBundle{}, ErrTimeout
	case msg := <-c.c.Incoming:

		for _, i := range valid {
			if i == msg.ID {
				return msg, nil
			}
		}

		return msg, ErrUnexpectedMessage
	}

	//return &ducktape.DuckTapeBundle{}, ErrUnexpectedMessage
}
