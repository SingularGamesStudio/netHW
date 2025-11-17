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
	role := flag.String("role", "", "")
	ip := flag.String("ip", "", "")
	cert := flag.String("cert", "", "")
	key := flag.String("key", "", "")
	flag.Parse()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		panic("graceful shutdown")
	}()

	var err error
	switch strings.ToLower(*role) {
	case "client":
		err = internal.TCPClient(*ip, *cert, *key)
	case "server":
		err = internal.TCPServer(*ip, *cert, *key)
	default:
		panic("role is neither client nor server")
	}
	if err != nil {
		panic(err)
	}
}
