package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
)

func main() {

	var serverAddr string
	flag.StringVar(&serverAddr, "server", ":8080", "server address")

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	quit := make(chan struct{})

	var roomID uint64
	buff := make([]byte, 8)
	for {
		fmt.Print("input roomID: ")
		n, err := fmt.Scanln(&roomID)
		if err != nil {
			fmt.Println("input error: ", err, "n: ", n)
			continue
		}
		if roomID != 0 {
			binary.LittleEndian.PutUint64(buff, roomID)
			conn.Write(buff)
			break
		}

	}

	go readMessageLoop(conn, quit)
	go writeMessageLoop(conn, roomID, quit)

	<-quit
}

func readMessageLoop(conn net.Conn, quit chan struct{}) {
	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if err == io.EOF {
				fmt.Println("server closed")
				quit <- struct{}{}
				return
			}
			fmt.Println("read error: ", err)
			continue
		}
		fmt.Println("read: ", string(buff[:n]))
	}
}

func writeMessageLoop(conn net.Conn, roomID uint64, quit chan struct{}) {
	var msg string
	for {
		fmt.Print("input message: ")
		fmt.Scanln(&msg)

		buff := make([]byte, len(msg))
		copy(buff, msg)

		if msg == "quit" {
			quit <- struct{}{}
			return
		}
		conn.Write(buff)
	}
}
