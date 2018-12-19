package main

import (
	"errors"
	"strings"
	"fmt"
	"net"
	"bufio"
	"os"
)

var nickname string
var serverAdress string
var serverName string
var userName []string

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the server adress and port (0.0.0.0:0000): ")
	serverAdress, err := reader.ReadString('\n')
	serverAdress = strings.TrimSuffix(serverAdress, "\n")

    // connection to localhost on port 2000
    conn, err := net.Dial("tcp", "127.0.0.1:2000")
    if err != nil {
        panic(err)
    }

    inputChan := make(chan string)
    receiveChan := make(chan string)

	fmt.Println("\nEnter Your Nickname: ")
	nickname, err := reader.ReadString('\n')
	conn.Write([]byte("TCCHAT_REGISTER\t"+nickname+"\n"))

	go receiveInput(inputChan)
	go receiveMessages(conn, receiveChan)


    for {
        select {
        case inputMessage := <-inputChan:
            //fmt.Println("SEND '", inputMessage, "' to server")
            conn.Write([]byte(inputMessage + "\n"))
        case receivedMessage := <-receiveChan:
            //fmt.Println("RECEIVED '", receivedMessage, "' from server")
			err := client_react (receivedMessage);
			if err != nil {
		        panic(err)
		    }
        }
    }
}

func receiveInput(inputChan chan string) {
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("Text to send: ")
        text, err := reader.ReadString('\n')
		if err != nil {
            panic(err)
        }
        text = strings.TrimSuffix(text, "\n")
		text = "TCCHAT_MESSAGE\t"+text
        inputChan <- text
    }
}

func receiveMessages(conn net.Conn, receiveChan chan string) {
    reader := bufio.NewReader(conn)
    for {
        text, err := reader.ReadString('\n')
        if err != nil {
            panic(err)
        }
        receiveChan <- text
    }
}

// Cl_Serv_react

func client_react(message string) error {

	var msgPieces []string
	typeMsg := ""
	argMsg1 := ""
	argMsg2 := ""

	msgPieces = strings.SplitN(message, "\n",1)
	msgPieces = strings.SplitN(msgPieces[0], "\t",3)

	if len(msgPieces) < 2 {
		return  errors.New("Not enough message's arguments");
	}

	typeMsg = msgPieces[0]
	argMsg1 = msgPieces[1]
	if len(msgPieces) > 2 {
		argMsg2 = msgPieces [2]
	}

	switch typeMsg {
	case "TCCHAT_WELCOME":
		welcome(argMsg1)
	case "TCCHAT_USERIN":
		userin(argMsg1)
	case "TCCHAT_USEROUT":
		userout(argMsg1)
	case "TCCHAT_BCAST":
		if argMsg2 == "" {
			return  errors.New("Empty message");
		} else if len(argMsg2) > 140 {
			return  errors.New("Message Payload over 140 character");
		}
		newMessage(argMsg1,argMsg2)
	default :
		return  errors.New("Undefined Type of message");
	}

	return nil
}

func welcome(nom_serv string) {
	fmt.Println("connecté au serveur :", nom_serv)
	serverName = nom_serv
}

func userin (nom_user string) {
	fmt.Println(nom_user, "rejoind le serveur")
	userName = append(userName, nom_user)
}

func userout (nom_user string) {
	fmt.Println(nom_user, "est OUT #Micdrop")
	for i := 0; i<len(userName); i++ {
		if (userName[i] == nom_user) {
			i = len(nom_user)
			userName[i] = userName[len(userName)-1]
			userName = userName[:len(userName)-1]
		}
	}
}

func newMessage (nom_user string, message string) {
	fmt.Println(nom_user, ":", message)
}
