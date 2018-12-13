//go function that react to the client
package main

import (
	"errors"
	"strings"
	"fmt"
	"net"
	"bufio"
)

users := make(map[string]string)

func serv_react(message string, ip string) error {
	var msgPieces []string
	typeMsg := ""
	argMsg := ""

	msgPieces = strings.SplitN(message, "\n",1)
	msgPieces = strings.SplitN(msgPieces[0], "\t",2)

	typeMsg = msgPieces[0]

	if len(msgPieces) == 2 {
		argMsg = msgPieces [1]
	}

	switch typeMsg {
	case "TCCHAT_REGISTER":
		msgPieces = strings.Split(argMsg, "\t") //no \t in nicknames
		if msgPieces[0] == "" {
			return  errors.New("no empty Nickname allowed");
		}
		registerUser (msgPieces[0],ip);
	case "TCCHAT_MESSAGE" :
		if len(argMsg) > 140 {
			return  errors.New("Message Payload over 140 character");
		}
		broadcast (argMsg)
	case "TCCHAT_DISCONNECT":
		disconnect (ip)
		default :
		var err error
		err = errors.New("Undefined Type of message")
		return err;
	}

	return nil
}





func registerUser (nickname string, ip string) {
	_,isYetConnect := users[ip]
	if (isYetConnect) {
		disconnect(ip)
	}else {
		users[ip] = nickname
		msg := "TCCHAT_USERIN\t"+nickname+"\n"
		sendMessage (msg,"")
	}
}

func broadcast (msg string) {
	msg := "TCCHAT_BCAST\t"+nickname+"\t"+msg+"\n"
	sendMessage (msg,"")
}

func disconnect (ip string) {
	nickname := users[ip]
	msg := "TCCHAT_USEROUT\t"+nickname+"\n"
	sendMessage (msg,"")
	delete (users, ip)

}

func sendMessage (message string, ip string) {
	// A la place d'utiliser cette methode, envoyer le message dans un channel
}

func main () {
	message := "TCCHAT_MESSAGE\tDamon\tje suis un message avec \t une tabulation (\\t)\n"
	ip := "127.0.0.0"
	err := serv_react (message, ip)
	fmt.Println (err)
}
