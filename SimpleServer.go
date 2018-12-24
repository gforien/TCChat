package main

import (
    "net"
    "fmt"
    "bufio"
    "strings"
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

        case onConnection, ok := <-connectChan:
                        if !ok {
                fmt.Println("Channel is closed !")
                break
            }
            fmt.Println("NEW CONN")
//            go sendMessage(Client{conn: onConnection, name : "undefined"}, "TCCHAT_WELCOME\t"+serverName)
            go getMsg(onConnection, msgChan)

        case onMessage, ok := <-msgChan:
                        if !ok {
                fmt.Println("Channel is closed !")
                break
            }
            fmt.Println("NEW MSG: ", onMessage)


        }
    }
}

func getMsg(conn net.Conn, msgChan chan string) {
    var msgPieces []string
    reader := bufio.NewReader(conn)

    for {
        fmt.Println("Listening on user "+conn.RemoteAddr().String())
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
            registerUser(conn, msgPieces[1]);
            conn.Write([]byte("TCCHAT_WELCOME\tLe chat de TC\n"))

        case "TCCHAT_MESSAGE":
            if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                panic("Error: Received message doesn't respect TC-Chat protocol.")
            }
            msgChan <- "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]

        case "TCCHAT_DISCONNECT":
            fmt.Println("TCCHAT_DISCONNECT")
        }
    }
}

func getInput() {
}

func getConn(listener net.Listener) {
    for {
        conn, err := listener.Accept()
        if err != nil {
            panic(err)
        }

        go func() {connectChan <- conn}()
    }
}

/* Users/Messages - related function */

func registerUser(connReceived net.Conn, nameReceived string) {
    aConn = append(aConn, Client{conn: connReceived, name: nameReceived})
    msgChan <- "TCCHAT_USERIN\t"+nameReceived
    //userInMessage := "TCCHAT_USERIN\t"+nameReceived
    //broadcastChan <- userInMessage
}

func sendMessage(client Client, msg string) {
    client.conn.Write([]byte(msg + "\n"))
}