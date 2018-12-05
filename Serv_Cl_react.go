//go function that react to the client
package main

import (
  "errors"
  "strings"
  "fmt"
)

func serv_react(message string, ip int) error {

  var msgPieces []string
  msgPieces = strings.SplitN(message, "\t",2)

  var typeMsg string
  typeMsg = msgPieces[0]

  msgPieces = strings.SplitN(msgPieces[1], "\n",1)

  var argMsg string
  argMsg = msgPieces[0]

  switch typeMsg {
  case "TCCHAT_REGISTER":
    registerUser (argMsg,ip);
  case "TCCHAT_MESSAGE" :
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

func registerUser (nickname string, ip int) {
  fmt.Println (ip, "est connecté avec le nom :", nickname)
}

func broadcast (msg string) {
  fmt.Println (" BROADCAST :", msg)
}

func disconnect (ip int) {
  fmt.Println ("disconnect ", ip)
}

func main () {
	arg1 := "TCCHAT_DISCONNECT\tDamon!\n"
	arg2 := 127000
	serv_react (arg1, arg2)
}
