package main

import (
	"net"
	"fmt"
	"bufio"
	"strings"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	aConn = make([]Client, 0)
	serverName = "TCChat Server"
	// channels definition
	connectChan = make(chan net.Conn)
	broadcastChan = make(chan string)

	invalidProtocol = "Received message doesn't respect TC-Chat protocol."
)

func main() {

	fmt.Println("Launching server...")
	listener, err := net.Listen("tcp", "127.0.0.1:2000")
	if err != nil {
		panic(err)
	}

	go getConn(listener)

	// mainloop
	for {
		select {
		case onConnection := <-connectChan:
			fmt.Println("NEW CONN")
			go sendMessage(Client{conn: onConnection, name : "undefined"}, "TCCHAT_WELCOME\t"+serverName)
			go getMsg(onConnection)

		case onBroadcast := <-broadcastChan:
			fmt.Println("BROADCAST : "+onBroadcast)
			for i := 0; i<len(aConn); i++ {
				go sendMessage(aConn[i], onBroadcast)
			}
		}
	}
}

func getMsg(conn net.Conn) {
	var msgPieces []string
	reader := bufio.NewReader(conn)

	isConnected := true
	defer func() {
		if r := recover(); r != nil {
			disconnect(conn)
			isConnected = false
		}
	}()

	for isConnected{
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimSuffix(text, "\n")

		msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)
		if len(msgPieces) < 2 || msgPieces[0] == "" || msgPieces[1] == ""{
			msgPieces = make([]string, 1)
		}

		switch msgPieces[0] {

		case "TCCHAT_REGISTER":
			notYetConnect := !disconnect(conn)
			if notYetConnect {
				aConn = append(aConn, Client{conn: conn, name: msgPieces[1]})
				broadcastChan <-"TCCHAT_USERIN\t"+msgPieces[1]
			}

		case "TCCHAT_MESSAGE":
			if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
				fmt.Println(invalidProtocol)
			} else {
				broadcastChan <- "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]
			}

		case "TCCHAT_DISCONNECT":
			disconnect(conn)

		default : fmt.Println(invalidProtocol)
		}
	}
}

func getConn(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		connectChan <- conn
	}
}


func sendMessage(client Client, msg string) {
	client.conn.Write([]byte(msg + "\n"))
}


func disconnect (conn net.Conn) bool{
	for i := 0; i<len(aConn); i++ {
		if conn == aConn[i].conn {
			broadcastChan <- "TCCHAT_USEROUT\t"+aConn[i].name
			aConn[i].conn.Close()
			aConn = append(aConn[:i], aConn[i+1:]...)
			return true
		}
	}
	return false
}
