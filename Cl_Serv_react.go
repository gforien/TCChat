//go function that react to the server
package main

import (
	"errors"
	"strings"
	"fmt"
)

func serv_react(message string, ip int) error {

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
}

func userin (nom_user string) {
	fmt.Println(nom_user, "rejoin le serveur")
}

func userout (nom_user string) {
	fmt.Println(nom_user, "est OUT #Micdrop")
}

func newMessage (nom_user string, message string) {
	fmt.Println(nom_user, ":", message)
}

func main () {
	message := "TCCHAT_WELCOME\t the lul-server \tje suis un message avec\tune tabulation (\\t)\n"
	ip := 127000
	err := serv_react (message, ip)
	fmt.Println (err)
}
