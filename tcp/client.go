package tcp

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	Conn net.Conn
	wg   sync.WaitGroup
}

func (c *Client) AddWaiting() {
	c.wg.Add(1)
}

func (c *Client) Done() {
	c.wg.Done()
}

func (c *Client) Close() error {
	ch := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(ch)
	}()

	select {
	case <-ch:
		return c.Conn.Close()
	case <-time.After(time.Second * 5):
		return c.Conn.Close()
	}
}
