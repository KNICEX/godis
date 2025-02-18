package tcp

import (
	"bufio"
	"context"
	"godis/pkg/logx"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type EchoHandler struct {
	closeChan chan struct{}
	closed    atomic.Bool

	connMap sync.Map // map[*Client]struct{}
	once    sync.Once
}

func NewEchoHandler() Handler {
	return &EchoHandler{
		closeChan: make(chan struct{}),
	}
}

func (e *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if e.closed.Load() {
		return
	}

	client := &Client{Conn: conn}
	e.connMap.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logx.L().Info("connection closed")
			} else {
				logx.L().Warn(err)
			}
			return
		}

		client.AddWaiting()
		b := []byte(msg)
		_, err = conn.Write(b)
		if err != nil {
			logx.L().Warn(err)
			return
		}
		client.Done()

	}
}

func (e *EchoHandler) Close() error {
	e.once.Do(func() {
		close(e.closeChan)
		e.closed.Store(true)
		wg := sync.WaitGroup{}
		e.connMap.Range(func(key, value interface{}) bool {
			client := key.(*Client)
			wg.Add(1)
			go func() {
				client.Close()
				wg.Done()
			}()
			return true
		})
		wg.Wait()
	})
	return nil
}
