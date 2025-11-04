//go:build linux

package internal

import (
	"fmt"
	"syscall"
)

func TCPClient(ip string) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	socketAddr, err := ipToSocketAddress(ip)
	if err != nil {
		return err
	}
	err = syscall.Connect(fd, socketAddr)
	if err != nil {
		return err
	}
	defer syscall.Shutdown(fd, syscall.SHUT_RDWR)
	fmt.Printf("Connected to %s\n", ip)

	chat(fd)
	return nil
}

func TCPServer(ip string) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)

	socketAddr, err := ipToSocketAddress(ip)
	if err != nil {
		return err
	}
	err = syscall.Bind(fd, socketAddr)
	if err != nil {
		return err
	}
	defer syscall.Shutdown(fd, syscall.SHUT_RDWR)

	if err := syscall.Listen(fd, 1); err != nil {
		return err
	}

	for {
		fmt.Printf("Awaiting client at %s\n", ip)
		clientFD, _, err := syscall.Accept(fd)
		if err != nil {
			return err
		}
		fmt.Println("Connected to client")

		chat(clientFD)

		// close connection to interrupt the other goroutine
		syscall.Shutdown(clientFD, syscall.SHUT_RDWR)
		syscall.Close(clientFD)

		fmt.Println("Client disconnected")
	}
}
