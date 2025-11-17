//go:build linux

package main

import (
	"context"
	"flag"
	"netHW/internal"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	protocol := flag.String("protocol", "", "")
	role := flag.String("role", "", "")
	ip := flag.String("ip", "", "")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		if strings.ToLower(*role) == "client" {
			cancel()
		} else {
			panic("shutdown")
		}
	}()

	var err error
	switch strings.ToLower(*protocol) {
	case "tcp":
		switch strings.ToLower(*role) {
		case "client":
			err = internal.TCPClient(ctx, *ip)
		case "server":
			err = internal.TCPServer(ctx, *ip)
		default:
			panic("role is neither client nor server")
		}
	case "udp":
		switch strings.ToLower(*role) {
		case "client":
			err = internal.UDPClient(ctx, *ip)
		case "server":
			err = internal.UDPServer(ctx, *ip)
		default:
			panic("role is neither client nor server")
		}
	default:
		panic("protocol is neither tcp nor udp")
	}
	if err != nil {
		panic(err)
	}
}
