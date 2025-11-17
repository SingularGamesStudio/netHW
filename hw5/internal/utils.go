package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

func chat(conn io.ReadWriter) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		buf := make([]byte, 100000)
		for {
			cnt, err := conn.Read(buf)
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
			err = write(conn, line)
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

func write(conn io.Writer, data string) error {
	b := []byte(data)
	for len(b) > 0 {
		cnt, err := conn.Write(b)
		if err != nil {
			return err
		}
		b = b[cnt:]
	}
	return nil
}
