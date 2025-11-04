//go:build linux

package main

import (
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

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		panic("graceful shutdown")
	}()

	var err error
	switch strings.ToLower(*protocol) {
	case "tcp":
		switch strings.ToLower(*role) {
		case "client":
			err = internal.TCPClient(*ip)
		case "server":
			err = internal.TCPServer(*ip)
		default:
			panic("role is neither client nor server")
		}
	case "udp":
		switch strings.ToLower(*role) {
		case "client":
			err = internal.UDPClient(*ip)
		case "server":
			err = internal.UDPServer(*ip)
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
