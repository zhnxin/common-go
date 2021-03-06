package common

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func NewUnixSocketServer(sockerFile string) *UnixSocketServer {
	s := &UnixSocketServer{
		sockerFile: sockerFile,
		signalHandlerFunc: map[string]DefaultUnixSocketHandler{
			"status": func(s *UnixSocketServer, c net.Conn, _ ...string) error {
				_, err := fmt.Fprintf(c, "running:%d", os.Getpid())
				return err
			},
		},
	}
	s.socketHandler = s.defaultsocketHandler
	s.SetSignalHandlerFunc("stop", func(is *UnixSocketServer, c net.Conn, _ ...string) error {
		_, err := fmt.Fprintf(c, "stop:%d", os.Getpid())
		is.cannel()
		return err
	})
	return s
}

func NewUnixSocketServerWith(sockerFile string, socketHandler UnixSocketHandler) *UnixSocketServer {
	return &UnixSocketServer{
		sockerFile:    sockerFile,
		socketHandler: socketHandler,
	}
}

type UnixSocketHandler func(*UnixSocketServer, net.Conn) error
type DefaultUnixSocketHandler func(*UnixSocketServer, net.Conn, ...string) error

type UnixSocketServer struct {
	sockerFile        string
	cannel            context.CancelFunc
	socketHandler     UnixSocketHandler
	signalHandlerFunc map[string]DefaultUnixSocketHandler
	stopCtx           context.Context
}

func (s *UnixSocketServer) defaultsocketHandler(_ *UnixSocketServer, c net.Conn) error {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return nil
		}

		data := buf[0:nr]
		signals := strings.Split(string(data), " ")
		fn, ok := s.signalHandlerFunc[signals[0]]
		if ok {
			err = fn(s, c, signals[1:]...)
		} else {
			_, err = fmt.Fprintf(c, "invalid signal")
		}
		if err != nil {
			return err
		}
	}
}
func (s *UnixSocketServer) SetSignalHandlerFunc(signal string, fn DefaultUnixSocketHandler) {
	if s.signalHandlerFunc == nil {
		panic("SetSignalHandlerFunc is diabled while using custom socketHandler")
	}
	s.signalHandlerFunc[signal] = fn
}

func (s *UnixSocketServer) Listen() (err error) {
	ln, err := net.Listen("unix", s.sockerFile)
	if err != nil {
		return err
	}
	var stopCannel context.CancelFunc
	s.stopCtx, stopCannel = context.WithCancel(context.Background())
	defer stopCannel()
	ctx, cannel := context.WithCancel(context.Background())
	s.cannel = cannel
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer ln.Close()
		defer stopCannel()
		select {
		case <-ctx.Done():
			fmt.Printf("shutting down by .Stop()\n")
		case sig := <-sigc:
			fmt.Printf("Caught signal %s: shutting down.\n", sig)
		}
	}()

	for {
		fd, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("Accept error:%v", err)
		}
		if err = s.socketHandler(s, fd); err != nil {
			return fmt.Errorf("socket handler:%v", err)
		}
	}
}

func (s *UnixSocketServer) Dial(cmd string) ([]byte, error) {
	c, err := net.Dial("unix", s.sockerFile)
	if err != nil {
		return nil, fmt.Errorf("Dial error:%v", err)
	}
	defer c.Close()
	if _, err := c.Write([]byte(cmd)); err != nil {
		return nil, err
	}
	return s.readReplay(c)
}

func (s *UnixSocketServer) readReplay(r io.Reader) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := r.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return buf[0:n], nil
}

//Stop the socket listener
func (s *UnixSocketServer) Stop() {
	if s.cannel != nil {
		s.cannel()
	}
	<-s.Stoped()
}

func (s *UnixSocketServer) Stoped() <-chan struct{} {
	if s.stopCtx == nil {
		return nil
	}
	return s.stopCtx.Done()
}

func (s *UnixSocketServer) SetHandler(handler UnixSocketHandler) {
	s.socketHandler = handler
}
