//Simple Client, you have to manually implement the protocol

package main

import (
	"fmt"
	"net"
	"bufio"
	"os"
)

var (
	inputChan = make(chan string)
	receiveChan = make(chan string)
	conn net.Conn
	connectionError error
	invalidProtocol = "Received message doesn't respect TC-Chat protocol."
)

func main() {

	//connecting to the server
	conn, connectionError = net.Dial("tcp", "127.0.0.1:2000")
	if connectionError != nil {panic(connectionError)}

	// launch the management of sent messages and receive messages
	go getMsg()
	go getInput()

	// displaying messages loop
	for {
		select {
		case input, ok := <-inputChan:
			if !ok {
				fmt.Println("Channel is closed !")
				break
			}
			fmt.Print("\nsending : "+input)
			_, err := conn.Write([]byte(input ))
			if err != nil {panic (err)}

		case msg, ok := <-receiveChan:
			if !ok {
				fmt.Println("Channel is closed !")
				break
			}
			fmt.Print("\nreceiving : "+msg)
		}
	}
}


func getInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		inputChan <- text
	}
}


func getMsg() {
	reader := bufio.NewReader(conn)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		receiveChan <- text
	}
}
