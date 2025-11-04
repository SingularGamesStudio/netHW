//go:build linux

package internal

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func chat(fd int) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		buf := make([]byte, 100000)
		for {
			cnt, err := syscall.Read(fd, buf)
			if cnt > 0 {
				if string(buf[:cnt]) == "DISCONNECT\n" {
					fmt.Println("DISCONNECT received")
					break
				}
				fmt.Print("message: ", string(buf[:cnt]))
			}
			if err != nil {
				fmt.Println("Error: ", err)
				break
			}
			if ctx.Err() != nil {
				break
			}
		}
		cancel()
	}()
	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				fmt.Println("Error: ", err)
				break
			}
			err = write(fd, line)
			if err != nil {
				fmt.Println("Error: ", err)
				break
			}
			if ctx.Err() != nil {
				break
			}
			if line == "DISCONNECT\n" {
				break
			}
		}
		cancel()
	}()

	<-ctx.Done()
}

func ipToSocketAddress(addr string) (*syscall.SockaddrInet4, error) {
	ip := net.ParseIP(strings.Split(addr, ":")[0])
	port, err := strconv.Atoi(strings.Split(addr, ":")[1])
	if err != nil {
		return nil, err
	}
	return &syscall.SockaddrInet4{
		Addr: [4]byte(ip.To4()),
		Port: port,
	}, nil
}
func write(fd int, data string) error {
	b := []byte(data)
	for len(b) > 0 {
		cnt, err := syscall.Write(fd, b)
		if err != nil {
			return err
		}
		b = b[cnt:]
	}
	return nil
}
