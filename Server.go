package main

import (
	"net"
	"fmt"
	"bufio"
	"strings"
	"os"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	aConn = make([]Client, 0)
	serverName string
	// channels definition
	connectChan = make(chan net.Conn)
	broadcastChan = make(chan string)
	writeLog = make(chan string)

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

	// Enter serverName
	fmt.Println("\nEnter The Server Name: ")
	str, err := reader.ReadString('\n')
	serverName = str
	if err != nil {panic(err)}
	serverName = strings.TrimSuffix(serverName, "\n")
	if "" == serverName {serverName = "TCChat_Server"}

	//create a file for displaying the logs
	f, errFile := os.Create("/tmp/TCChat_"+serverName) // acces the file with : tail -f /tmp/TCChat_[serverName]
	if errFile != nil {panic(errFile)}

	//Launching the server
	fmt.Println("Launching server : "+serverName)
	listener, err := net.Listen("tcp", serverAdress)
	if err != nil {
		panic(err)
	}

	go getConn(listener) //listen to new connection
	go getInput() //waiting for input

	// mainloop
	for {
		select {
		case onConnection := <-connectChan:
			_ , err := f.WriteString("NEW CONNECTION"+"\n")
			if err != nil {panic(err)}
			go sendMessage(Client{conn: onConnection, name : "undefined"}, "TCCHAT_WELCOME\t"+serverName)
			go getMsg(onConnection)

		case onBroadcast := <-broadcastChan:
			_ , err := f.WriteString("BROADCAST : "+onBroadcast+"\n")
			if err != nil {panic(err)}
			for i := 0; i<len(aConn); i++ {
				go sendMessage(aConn[i], onBroadcast)
			}

		case onLog := <- writeLog :
			_ , err := f.WriteString(onLog+"\n")
			if err != nil {panic(err)}
		}
	}
}

// handle message from a given client
func getMsg(conn net.Conn) {

	nickname := "undefined"
	var msgPieces []string
	reader := bufio.NewReader(conn)

	// if the method panic the loop condition become false, goroutine stop
	isConnected := true
	defer func() {
		if r := recover(); r != nil {
			disconnect(conn)
			isConnected = false
		}
	}()

	// handling the messages
	for isConnected{
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimSuffix(text, "\n")

		msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)

		switch msgPieces[0] {

		case "TCCHAT_REGISTER":
			if len(msgPieces) < 2 || msgPieces[1] == "" {
				writeLog <- invalidProtocol
			}
			notYetConnect := !disconnect(conn)
			if notYetConnect {
				nickname = msgPieces[1]
				aConn = append(aConn, Client{conn: conn, name: msgPieces[1]})
				broadcastChan <-"TCCHAT_USERIN\t"+msgPieces[1]
			}

		case "TCCHAT_MESSAGE":
			if len(msgPieces) < 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
				writeLog <- invalidProtocol
			} else {
				broadcastChan <- "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]
			}

		case "TCCHAT_DISCONNECT":
			disconnect(conn)

		case "TCCHAT_USERS" :
			for i := 0; i<len(aConn); i++ {
				if nickname == aConn[i].name {
					go sendMessage(aConn[i],"TCCHAT_USERLIST\t"+strings.Replace(giveUsers(),"\n","\r",-1))
					i=len(aConn)
				}
			}

		case "TCCHAT_TELL" :
			if len(msgPieces) < 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
				writeLog <- invalidProtocol
			} else {
				for i := 0; i<len(aConn); i++ {
					if msgPieces[1] == aConn[i].name {
						go sendMessage(aConn[i],"TCCHAT_PRIVATE\t"+nickname+"\t"+msgPieces[2])
						i=len(aConn)
					}
				}
			}

		default : writeLog <- invalidProtocol
		}
	}
}

// handle new client connection
func getConn(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		connectChan <- conn
	}
}

// handle a input stream for the server
func getInput () {
	for {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {panic(err)}
		input = strings.TrimSuffix(input, "\n")
		writeLog <- "INPUT : "+input
		msgPieces := strings.SplitN(input," ",2)

		switch msgPieces[0] {

		case "/broadcast" :
			broadcastChan <- "TCCHAT_BCAST\t"+serverName+"\t"+msgPieces[1]

		case "/tell" :
			msgPieces = strings.SplitN(msgPieces[1]," ",2)
			for i := 0; i<len(aConn); i++ {
				if msgPieces[0] == aConn[i].name {
					go sendMessage(aConn[i],"TCCHAT_PRIVATE\t"+serverName+"\t"+msgPieces[1])
					i=len(aConn)
				}
			}

		case "/disconnect" :
			for i := 0; i<len(aConn); i++ {
				if msgPieces[1] == aConn[i].name {
					aConn[i].conn.Close()
				}
			}

		case "/users" :
			writeLog <- "\n"+giveUsers()
		}
	}
}

// other function :

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

func giveUsers () string {
	var str string
	for i := 0; i<len(aConn); i++ {
		str += aConn[i].name + "\n"
	}
	return str
}
