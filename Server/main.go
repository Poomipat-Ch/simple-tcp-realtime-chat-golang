package main

import (
	"flag"
	"log"
)

func main() {

	var listenAddr string
	flag.StringVar(&listenAddr, "l", ":8080", "listen address")
	flag.Parse()

	s := NewServer(listenAddr)

	go func() {
		for msg := range s.msgchan {
			room := getRoomFromRemoteAddr(s.rooms, msg.remoteAddr.String())
			for peer := range room.peerConn {
				if peer != msg.remoteAddr.String() {
					room.peerConn[peer].Write(msg.Msg)
				}
			}
		}
	}()

	log.Fatal(s.Start())
}
