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
	inputChan = make(chan string)
	receiveChan = make(chan string)
	conn net.Conn
	connectionError error
	invalidProtocol = "Received message doesn't respect TC-Chat protocol."
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Enter server address
	fmt.Println("Enter the server adress and port (0.0.0.0:0000): ")
	serverAdress, err := reader.ReadString('\n')
	if err != nil {panic(err)}
	serverAdress = strings.TrimSuffix(serverAdress, "\n")
	if "" == serverAdress {serverAdress = "127.0.0.1:2000"}

	// Enter nickname
	fmt.Println("\nEnter Your Nickname: ")
	str, err := reader.ReadString('\n')
	if err != nil {panic(err)}
	nickname = str
	nickname = strings.TrimSuffix(nickname, "\n")
	if "" == nickname {nickname = "defaultName"}

	//create a file for displaying server's message
	f, errFile := os.Create("/tmp/TCChat_"+nickname) // acces the file with : tail -f /tmp/TCChat_[nickname]
	if errFile != nil {panic(errFile)}

	//connecting to the server
	conn, connectionError = net.Dial("tcp", serverAdress)
	if connectionError != nil {panic(connectionError)}

	// launch the management of sent messages and receive messages
	go getMsg()
	go getInput()

	// Send first message
	_ ,errCo := conn.Write([]byte("TCCHAT_REGISTER\t"+nickname + "\n"))
	if errCo != nil {panic(errCo)}

	// displaying messages loop
	for {
		select {
		case input, ok := <-inputChan:
			if !ok {
				fmt.Println("Channel is closed !")
				break
			}
			_, err := conn.Write([]byte(input + "\n"))
			if err != nil {panic (err)}

		case msg, ok := <-receiveChan:
			if !ok {
				fmt.Println("Channel is closed !")
				break
			}
			_ , err := f.WriteString(msg+"\n")
			if err != nil {panic(err)}
		}
	}
}


func getInput() {
	var msgPieces []string
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		text = strings.TrimSuffix(text, "\n")

		if strings.HasPrefix(text,"/") {
			msgPieces = strings.SplitN(text, " ", 2)
			switch msgPieces[0] {
			case "/help" : receiveChan <- "/help : print this help page\n/disconnect : close the client\n/users : print the list of connected users\n/mp <recipient>\t<message_payload> : send a private message to a given recipient"
			case "/disconnect" : inputChan <- "TCCHAT_DISCONNECT"
			case "/mp" :
				msgPieces = strings.SplitN(msgPieces[1], "\t", 2)
				inputChan <- "TCCHAT_TELL\t"+msgPieces[0]+"\t"+msgPieces[1]
			case "/users" : inputChan <- "TCCHAT_USERS"
			default : receiveChan <- "Undefined command. Try /help to see available ones."
			}
		}else {
			inputChan <- "TCCHAT_MESSAGE\t"+nickname+"\t"+text
		}
	}
}


func getMsg() {
	var msgPieces []string
	reader := bufio.NewReader(conn)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		text = strings.TrimSuffix(text, "\n")
		msgPieces = strings.SplitN(text, "\t", 3)

		if len(msgPieces) < 2 || msgPieces[1] == "" {msgPieces = make([]string, 1)}

		switch msgPieces[0] {
			case "TCCHAT_WELCOME" : receiveChan <- "Welcome on the server : " + msgPieces [1]
			case "TCCHAT_USERIN" : receiveChan <- "User in : " + strings.Split(msgPieces[1], "\n")[0]
			case "TCCHAT_USEROUT" : receiveChan <- "User out : " + msgPieces [1]
			case "TCCHAT_USERLIST" : receiveChan <- strings.Replace(msgPieces[1],"\r","\n",-1)
			case "TCCHAT_BCAST":
				if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
					fmt.Println(invalidProtocol)
				}else {
					receiveChan <- msgPieces[1]+" say : "+msgPieces[2]
				}
			case "TCCHAT_PRIVATE" :
				if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
					fmt.Println(invalidProtocol)
				}else {
					receiveChan <- msgPieces[1]+" tell : "+msgPieces[2]
				}
			default : fmt.Println(invalidProtocol)
		}
	}
}
