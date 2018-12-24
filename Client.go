package main

import (
	"strings"
	"fmt"
	"net"
	"bufio"
	"os"
	"errors"
)

var (
	inputChan = make(chan string)
	receiveChan = make(chan string)
	conn net.Conn
	connectionError error
	protocolError = errors.New("Received message doesn't respect TC-Chat protocol.")
	isConnected bool
)

func main() {
	isConnected = true
	reader := bufio.NewReader(os.Stdin)

	// Enter server address
	fmt.Println("Enter the server adress and port (0.0.0.0:0000): ")
	serverAdress, err := reader.ReadString('\n')
	if err != nil {panic(err)}
	serverAdress = strings.TrimSuffix(serverAdress, "\n")
	if "" == serverAdress {serverAdress = "127.0.0.1:2000"}

	// Enter nickname
	fmt.Println("\nEnter Your Nickname: ")
	nickname, err := reader.ReadString('\n')
	if err != nil {panic(err)}
	nickname = strings.TrimSuffix(nickname, "\n")
	if "" == nickname {nickname = "defaultName"}

	//create a file for displaying server's message
	f, errFile := os.Create("/tmp/TCChat_"+nickname) // acces the file with : tail -f /tmp/TCChat_[nickname]
	if errFile != nil {panic(err)}

	//connecting to the server
	conn, connectionError = net.Dial("tcp", serverAdress)
	if connectionError != nil {panic(connectionError)}

	// launch the management of sent messages and receive messages
	go getMsg()
	go getInput(nickname)

	// Send first message
	_ ,errCo := conn.Write([]byte("TCCHAT_REGISTER\t"+nickname + "\n"))
	if errCo != nil {panic(errCo)}

	// displaying messages
	for isConnected {
		select {
		case input, ok := <-inputChan:
			if !ok {
				fmt.Println("Channel is closed !")
				break
			}
			fmt.Println("sending : "+input) // this print is for debugging, our own message is display with the reception of TCCHAT_BCAST
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


func getInput(nickname string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		text = strings.TrimSuffix(text, "\n")
		if "DISCONNECT" == text {
			inputChan <- "TCCHAT_DISCONNECT\t"+nickname
			// l'arret ce fait lors de l'erreur provoqué par la fermeture de la co par le serveur
		} else {
			inputChan <- "TCCHAT_MESSAGE\t"+nickname+"\t"+text
		}
	}
}


func getMsg() {
	var msgPieces []string
	reader := bufio.NewReader(conn)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimSuffix(text, "\n")

		msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)

		if len(msgPieces) < 2 || msgPieces[0] == "" || msgPieces[1] == "" {panic(protocolError)}

		switch msgPieces[0] {
			case "TCCHAT_WELCOME":
				receiveChan <- "Welcome on the server : " + msgPieces [1]

			case "TCCHAT_USERIN":
				receiveChan <- "User in : " + msgPieces [1]

			case "TCCHAT_USEROUT":
				receiveChan <- "User out : " + msgPieces [1]

			case "TCCHAT_BCAST":
				if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {panic(protocolError)}
				receiveChan <- msgPieces[1]+" say : "+msgPieces[2]

			default : panic(protocolError)
		}
	}
}
