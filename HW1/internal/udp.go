//go:build linux

package internal

import (
	"context"
	"fmt"
	"syscall"
)

func UDPClient(ctx context.Context, ip string) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM|syscall.SOCK_CLOEXEC, syscall.IPPROTO_UDP)
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
	fmt.Printf("Connected to %s\n", ip)
	defer func() {
		_ = write(fd, "DISCONNECT\n")
	}()

	chat(ctx, fd)
	return nil
}

func UDPServer(ctx context.Context, ip string) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM|syscall.SOCK_CLOEXEC, syscall.IPPROTO_UDP)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}
	socketAddr, err := ipToSocketAddress(ip)
	if err != nil {
		return err
	}
	err = syscall.Bind(fd, socketAddr)
	if err != nil {
		return err
	}

	for {
		fmt.Printf("Awaiting client at %s\n", ip)
		buf := make([]byte, 1)
		var clientAddr syscall.Sockaddr
		var cnt int
		for {
			cnt, clientAddr, err = syscall.Recvfrom(fd, buf, 0)
			if err != nil {
				return err
			}
			if cnt > 0 {
				break
			}
		}
		if err := syscall.Connect(fd, clientAddr); err != nil {
			return err
		}
		fmt.Println("Connected to client")

		chat(ctx, fd)
		fmt.Println("Client disconnected")
	}
}

// go run ./... -protocol=udp -role=server -ip=192.168.0.1:7777
