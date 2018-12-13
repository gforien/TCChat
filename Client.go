package main

import "net"
import "fmt"
import "bufio"
import "os"

func main() {

    // connection to localhost on port 8081
    conn, err := net.Dial("tcp", "127.0.0.1:2000")
    if err != nil {
        panic(err)
    }

    inputChan := make(chan string)
    go receiveInput(inputChan)
    receiveChan := make(chan string)
    go receiveMessages(conn, receiveChan)

    for {
        select {
        case inputMessage := <-inputChan:
            fmt.Println("SEND '", inputMessage, "' to server")
            conn.Write([]byte(inputMessage + "\n"))
        case receivedMessage := <-receiveChan:
            fmt.Println("RECEIVED '", receivedMessage, "' from server")
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
