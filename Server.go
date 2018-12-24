package main

import (
	"net"
	"fmt"
	"bufio"
	"strings"
	"errors"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	aConn = make([]Client, 1)
	serverName = "TCChat Server"
	// channels definition
	inputChan = make(chan string)
	connectChan = make(chan net.Conn)
	msgChan = make(chan string)
	broadcastChan = make(chan string)

	protocolError = errors.New("Received message doesn't respect TC-Chat protocol.")
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

		case onMessage := <-msgChan:
			fmt.Println("NEW MSG: ", onMessage)

		case onBroadcast := <-broadcastChan:
			for _, client := range aConn {
				go sendMessage(client, onBroadcast)
			}

		case onInput := <-inputChan:
			fmt.Println("NEW INPUT: ", onInput)
		}
	}
}

func getMsg(conn net.Conn) {
	var msgPieces []string
	reader := bufio.NewReader(conn)

	for {
		text, err := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if err != nil {
			panic(err)
		}

		msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)
		if len(msgPieces) < 2 || msgPieces[0] == "" || msgPieces[1] == ""{
			panic("Error: Received message doesn't respect TC-Chat protocol.")
		}

		switch msgPieces[0] {

		case "TCCHAT_REGISTER":
			// TEST IF CLIENT EXIST YET (if true, disconnect the client, send UserOut)
			aConn = append(aConn, Client{conn: conn, name: msgPieces[1]})
			broadcastChan <-"TCCHAT_USERIN\t"+msgPieces[1]

		case "TCCHAT_MESSAGE":
			if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {panic(protocolError)}
			broadcastChan <- "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]

		case "TCCHAT_DISCONNECT":
			fmt.Println("TCCHAT_DISCONNECT")

		default panic (protocolError)
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
