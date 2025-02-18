package main

import (
	"context"
	"errors"
	"godis/pkg/logx"
	"godis/tcp"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Run(addr string, handler tcp.Handler) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer l.Close()
	log.Println("Listening on", addr)

	closeChan := make(chan struct{})
	signChan := make(chan os.Signal)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		<-signChan
		logx.L().Info("closing server")
		close(closeChan)
	}()

	ListenAndServe(l, handler, closeChan)
}

func ListenAndServe(l net.Listener, handler tcp.Handler, closeChan chan struct{}) {
	go func() {
		<-closeChan
		_ = l.Close()
		_ = handler.Close()
	}()

	wg := sync.WaitGroup{}
	for {
		conn, err := l.Accept()
		if err != nil {
			// 主动关闭
			if errors.Is(err, net.ErrClosed) {
				break
			}
			logx.L().Errorf("failed to accept: %v", err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.Handle(context.Background(), conn)
		}()
	}
	wg.Wait()
}

func main() {
	Run(":8888", tcp.NewEchoHandler())
}
