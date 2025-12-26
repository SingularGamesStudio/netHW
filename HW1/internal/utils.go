//go:build linux

package internal

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const bufSize = 100000

func chat(ctx context.Context, fd int) {
	ctx1, cancel := context.WithCancel(ctx)
	go func() {
		for {
			msg, err := readLine(fd)
			if msg == "DISCONNECT\n" {
				fmt.Println("DISCONNECT received")
				break
			}
			if err != nil {
				fmt.Println("Error: ", err)
				break
			}
			if ctx1.Err() != nil {
				break
			}
			fmt.Printf("message: %q\n", msg)
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
			if ctx1.Err() != nil {
				break
			}
			if line == "DISCONNECT\n" {
				break
			}
		}
		cancel()
	}()

	<-ctx1.Done()
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

func readLine(fd int) (string, error) {
	buf := make([]byte, bufSize)
	var lineBuf bytes.Buffer

	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			return "", err
		}
		if n == 0 {
			if lineBuf.Len() == 0 {
				return "", errors.New("connection closed")
			}
			return lineBuf.String(), nil
		}

		i := bytes.IndexByte(buf[:n], '\n')
		if i >= 0 {
			lineBuf.Write(buf[:i+1])
			return lineBuf.String(), nil
		}
		lineBuf.Write(buf[:n])
	}
}
