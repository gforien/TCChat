package main

import (
	"strings"
	"fmt"
	"net"
	"bufio"
	"os"
)

var (
	nickname string
	userName []string

	inputChan = make(chan string)
	receiveChan = make(chan string)
	conn net.Conn
	connectionError error
)

type Server struct {
	name string
}

func main() {

	/*    // Enter server address
	reader := bufio.NewReader(os.Stdin)
	//fmt.Println("Enter the server adress and port (0.0.0.0:0000): ")
	serverAdress, err := reader.ReadString('\n')
	serverAdress = strings.TrimSuffix(serverAdress, "\n")
	// Enter nickname
	fmt.Println("\nEnter Your Nickname: ")
	nickname, err := reader.ReadString('\n')
	if err != nil {
	panic(err)
}
// Connect to server
conn, err := net.Dial("tcp", serverAdress)
if err != nil {
panic(err)
}*/

nickname = "Gabriel"
conn, connectionError = net.Dial("tcp", "127.0.0.1:2000")
if connectionError != nil {
	panic(connectionError)
}

go getInput()
go getMsg()

// Send first message
go func() {inputChan <- "TCCHAT_REGISTER\t"+"Gabriel"}()

for {
	select {
	case input, ok := <-inputChan:
		if !ok {
			fmt.Println("Channel is closed !")
			break
		}
		conn.Write([]byte(input + "\n"))
		fmt.Println(input)
	case msg, ok := <-receiveChan:
		if !ok {
			fmt.Println("Channel is closed !")
			break
		}
		fmt.Println("NEW MSG :"+msg)
	}
}
}

func getInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Text to send: ")
		text, err := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if err != nil {
			panic(err)
		}
		inputChan <- "TCCHAT_MESSAGE\t"+nickname+"\t"+text
	}
}

func getMsg() {
	var msgPieces []string
	reader := bufio.NewReader(conn)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("text = "+text)
			panic(err)
		}
		text = strings.TrimSuffix(text, "\n")

		msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)
		if len(msgPieces) < 2 || msgPieces[0] == "" || msgPieces[1] == ""{
			fmt.Println(msgPieces)
			panic("Error: Received message doesn't respect TC-Chat protocol.")
		}

		switch msgPieces[0] {

		case "TCCHAT_WELCOME":
			receiveChan <- "WELCOME"

		case "TCCHAT_USERIN":
			receiveChan <- "USER IN"

		case "TCCHAT_USEROUT":
			receiveChan <- "USER OUT"

		case "TCCHAT_BCAST":
			if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
				panic("Error: Received message doesn't respect TC-Chat protocol.")
			}
			receiveChan <- "USER "+msgPieces[1]+" SAID "+msgPieces[2]
		}
	}
}
