package main

import (
    "net"
    "fmt"
    "flag"
    "bufio"
    "strings"
    "sync"
)

type Client struct {
    name string
}

func main() {
    var (
        userMap = make(map[net.Conn]*Client)    // map of Clients
        address *string                         // address and port of the server
        serverName *string                      // name of the server
        //msgChan = make(chan string)           // channel in which all message will be put
        mutex = &sync.Mutex{}
    )

    // address and server name are command line arguments
    address = flag.String("address", "127.0.0.1:2000", "IP address and port of the server")
    serverName = flag.String("name", "TC Community Server", "name of the server")
    flag.Parse()

    // launch server and panic if it doesn't work
    fmt.Println("main():\t\tLaunching server (name = "+*serverName+" ; address = "+*address+")")
    listener, connError := net.Listen("tcp", *address)
    if connError != nil {
        panic(connError)
    } else {
        fmt.Println("main():\t\tConnection established.")
    }

    // launch connection infinite loop, which will launch all other goroutines
    getConn(serverName, listener, userMap, mutex)
}


func getConn(serverName *string, listener net.Listener, userMap map[net.Conn]*Client,mutex *sync.Mutex) {
    fmt.Println("getConn():\tlistening on new connections.")

    for {
        // upon receiving a new user :
        conn, err := listener.Accept()
        fmt.Println("getConn():\tnew connection "+conn.RemoteAddr().String())
        if err != nil {
            fmt.Println("getConn():\terror when accepting new connection -> break\n"+err.Error())
            break
        }

        // 1) we update the user map
        mutex.Lock()
        userMap[conn] = &Client{name : "undefined"}
        fmt.Println("getConn():\tuserMap updated -> "+map2string(userMap))
        mutex.Unlock()

        // 2) we send WELCOME
        fmt.Println("getConn():\tsend message 'TCCHAT_WELCOME\t"+ *serverName+"' to <"+userMap[conn].name+">")
        _ , firstMessageErr := conn.Write([]byte("TCCHAT_WELCOME\t"+ *serverName +"\n"))
        if firstMessageErr != nil {
            fmt.Println("getConn():\terror when sending TCCHAT_WELCOME to <"+userMap[conn].name+"> -> break\n"+firstMessageErr.Error())
            break
        }

        // 3) we listen to him, because he's supposed to send REGISTER right after connecting
        go getMsg(serverName, conn, userMap,mutex)
    }
}


func getMsg(serverName *string, conn net.Conn, userMap map[net.Conn]*Client, mutex *sync.Mutex) {
    fmt.Println("\ngetMsg():\tlistening on messages from <"+userMap[conn].name+">")
    var msgPieces []string
    reader := bufio.NewReader(conn)

    for {
        // receiving and parsing the new message
        text, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("getMsg():\terror on reading message from <"+userMap[conn].name+"> ("+err.Error()+") -> force disconnect")
            go sendToAll(userMap, "TCCHAT_USEROUT\t"+ userMap[conn].name +"\n")
           
            
            fmt.Println("getMsg():\tUSEROUT\t"+ userMap[conn].name +" sent to all")
            //mutex.Lock()
            _, ok := userMap[conn]
            //mutex.Unlock()
            if ok {
				mutex.Lock()
                delete(userMap, conn)
                mutex.Unlock()
            } else {
                fmt.Println("getMsg():\terror on disconnecting <"+userMap[conn].name+"> ("+err.Error()+") -> break")
                break
            }
            fmt.Println("getMsg():\tuserMap updated -> "+map2string(userMap))
        }
        text = strings.TrimSuffix(text, "\n")
        msgPieces = strings.SplitN(strings.Split(text, "\n")[0], "\t", 3)
        if len(msgPieces) < 2 || msgPieces[0] == "" || msgPieces[1] == ""{
            fmt.Println("getMsg():\terror on parsing message from <"+userMap[conn].name+"> -> break")
            break
        }

        switch msgPieces[0] {

        // upon REGISTER, we send USERIN to all clients
        case "TCCHAT_REGISTER":
            fmt.Println("getMsg():\tgot 'REGISTER\t"+msgPieces[1]+"' from <"+userMap[conn].name+">")
            mutex.Lock()
            userMap[conn].name = msgPieces[1]
            mutex.Unlock()
            fmt.Println("getMsg():\tuserMap updated -> "+map2string(userMap))
            go sendToAll(userMap, "TCCHAT_USERIN\t"+ userMap[conn].name +"\n")
            fmt.Println("getMsg():\tUSERIN\t"+ userMap[conn].name +" sent to all")

        // upon MESSAGE, we send BCAST to all clients
        case "TCCHAT_MESSAGE":
            fmt.Println("getMsg():\tgot 'MESSAGE\t"+msgPieces[1]+"\t"+msgPieces[2]+"' from <"+userMap[conn].name+">")
            if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                fmt.Println("getMsg():\terror on parsing message (classed MESSAGE) from <"+userMap[conn].name+" -> break")
                break
            }
            go sendToAll(userMap, "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]+"\n")
            fmt.Println("getMsg():\tBCAST\t"+msgPieces[1]+"\t"+msgPieces[2]+" sent to all")

        // upon DISCONNECT, we delete the user from userMap and send USEROUT to all clients
        case "TCCHAT_DISCONNECT":
            fmt.Println("getMsg():\tDISCONNECT from <"+userMap[conn].name+">")
            go sendToAll(userMap, "TCCHAT_USEROUT\t"+ userMap[conn].name +"\n")
            fmt.Println("getMsg():\tUSEROUT\t"+ userMap[conn].name +" sent to all")
            //mutex.Lock()
            _, ok := userMap[conn]
            //mutex.Unlock()
            if ok {
				mutex.Lock()
                delete(userMap, conn)
                mutex.Unlock()
            }
            fmt.Println("getMsg():\tuserMap updated -> "+map2string(userMap))
        }
    }
}

func sendToAll(userMap map[net.Conn]*Client, wholeMessage string) {
    for userConn, userClient := range userMap {
        _ , err := userConn.Write([]byte(wholeMessage))
        if err != nil {
            msgCut := strings.Split(wholeMessage, "\n")[0]
            fmt.Println("sendToAll():\terror when sending '"+msgCut+"' to <"+userClient.name+"> : "+err.Error())
        }
    }
}

func map2string(userMap map[net.Conn]*Client) string {
    str := "[ "
    for key, val := range userMap {
        str += key.RemoteAddr().String()+":"+val.name+", "
    }
    return str + " ]"
}
