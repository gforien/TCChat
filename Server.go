package main

import "net"
import "fmt"
import "bufio"
import "strings"
//import "os"

func main() {

    fmt.Println("Launching server...")
    listener, err := net.Listen("tcp", "127.0.0.1:2000")
    if err != nil {
        panic(err)
    }

    inputChan := make(chan string)
    connectChan := make(chan net.Conn)
    msgChan := make(chan string)
    go getConn(listener, connectChan)

    for {
        select {
        case onConnection := <-connectChan:
            fmt.Println("NEW CONN")
            go getMsg(onConnection, msgChan)
        case onMessage := <-msgChan:
            fmt.Println("NEW MSG: ", onMessage)
        case onInput := <-inputChan:
            fmt.Println("NEW INPUT: ", onInput)
        }
    }
}

func getMsg(conn net.Conn, msgChan chan string) {
    reader := bufio.NewReader(conn)
    for {
        text, err := reader.ReadString('\n')
        text = strings.TrimSuffix(text, "\n")
        if err != nil {
            panic(err)
        }
        msgChan <- text
    }
}

func getInput(inputChan chan string) {}


func getConn(listener net.Listener, connectChan chan net.Conn) {
    for {
        conn, err := listener.Accept()
        if err != nil {
            panic(err)
        }

        connectChan <- conn
    }
}
