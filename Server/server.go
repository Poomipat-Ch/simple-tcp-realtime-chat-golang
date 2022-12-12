package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Room struct {
	roomID   uint64
	roomName string
	peerConn map[string]net.Conn
}

type Message struct {
	Msg        []byte
	remoteAddr net.Addr
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitchan   chan struct{}
	msgchan    chan Message
	rooms      map[uint64]*Room
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitchan:   make(chan struct{}),
		msgchan:    make(chan Message),
		rooms:      make(map[uint64]*Room),
	}
}

func (s *Server) Start() error {
	fmt.Printf("server start at %s\n", s.listenAddr)
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln
	go s.AcceptLoop()
	<-s.quitchan
	close(s.msgchan)
	return nil
}

func (s *Server) AcceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Printf("accept error: %s\n", err)
			continue
		}
		fmt.Printf("accept a new connection: %s\n", conn.RemoteAddr())
		ok := s.readRoomLoop(conn)
		if ok {
			go s.readLoop(conn)
		}
	}
}

func (s *Server) readRoomLoop(conn net.Conn) bool {
	buf := make([]byte, 8)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("connection closed by server: %s\n", conn.RemoteAddr())
				s.quitchan <- struct{}{}
			}
			fmt.Printf("[Read Room] read error: %s\n", err)
			return false
		}
		if n > 8 {
			fmt.Printf("[Read Room] wrong room id length: %d\n", n)
			conn.Write([]byte(fmt.Sprintf("wrong room id length expect 8 but receive %d", n)))
			continue
		}

		roomID := binary.LittleEndian.Uint64(buf[:n])
		fmt.Printf("[Read Room] read room id: %d buffer: %v\n", roomID, buf)
		room, ok := s.rooms[roomID]
		if !ok {
			room = &Room{
				roomID:   roomID,
				peerConn: make(map[string]net.Conn),
			}
			s.rooms[roomID] = room
		}
		room.peerConn[conn.RemoteAddr().String()] = conn
		break
	}
	return true
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 8+1024) // 8 bytes for room id, 1024 bytes for message
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("connection closed by client: %s\n", conn.RemoteAddr())
				return
			}
			fmt.Printf("read error: %s\n", err)
			return
		}
		message := Message{
			Msg:        buff[:n],
			remoteAddr: conn.RemoteAddr(),
		}
		s.msgchan <- message
	}
}

func getRoomFromRemoteAddr(rooms map[uint64]*Room, remoteAddr string) *Room {
	for _, room := range rooms {
		if _, ok := room.peerConn[remoteAddr]; ok {
			return room
		}
	}
	return nil
}
