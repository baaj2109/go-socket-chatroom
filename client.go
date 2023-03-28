package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("connect to ", conn.RemoteAddr())

	/// listen to server message
	go func() {
		buf := make([]byte, 5*1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			msg := string(buf[:n])
			fmt.Print(msg)
		}
	}()

	/// sent msg to server
	for {
		var msg string
		_, _ = fmt.Scanf("%s", &msg)
		msg = strings.Trim(msg, "\r\n")
		if "exit" == msg {
			fmt.Printf("close connect")
			conn.Close()
			return
		}
		_, _ = conn.Write([]byte(msg))
	}

}
