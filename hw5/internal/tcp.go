package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

func TCPClient(ip, rootCertFile string) error {
	tlsCfg := &tls.Config{}
	if path := os.Getenv("SSLKEYLOGFILE"); path != "" {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		tlsCfg.KeyLogWriter = f
	}

	// the server cert is self-signed, so we just pass it as trusted
	if rootCertFile != "" {
		rootCert, err := os.ReadFile(rootCertFile)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(rootCert) {
			return fmt.Errorf("failed to parse cert")
		}
		tlsCfg.RootCAs = pool
		tlsCfg.ServerName = "localhost"
	} else {
		tlsCfg.InsecureSkipVerify = true
	}

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		return err
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, tlsCfg)
	defer tlsConn.Close()

	if err := tlsConn.Handshake(); err != nil {
		return err
	}
	fmt.Printf("Connected to %s\n", ip)

	chat(tlsConn)
	return nil
}

func TCPServer(ip, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server, err := net.Listen("tcp", ip)
	if err != nil {
		return err
	}
	defer server.Close()

	for {
		fmt.Printf("Awaiting client at %s\n", ip)
		conn, err := server.Accept()
		if err != nil {
			return err
		}
		tlsConn := tls.Server(conn, tlsCfg)
		err = tlsConn.Handshake()
		if err != nil {
			fmt.Println("TLS handshake failed:", err)
		} else {
			fmt.Println("Connected to client")

			chat(tlsConn)
		}

		// close connection to interrupt the other goroutine
		tlsConn.Close()

		fmt.Println("Client disconnected")
	}
}
