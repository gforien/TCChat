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


func getConn(serverName *string, listener net.Listener, userMap map[net.Conn]*Client, mutex *sync.Mutex) {
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
    var wholeMessage string
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
        msgPieces = strings.SplitN(text, "\t", 4)
        if len(msgPieces) < 1 || msgPieces[0] == "" {
            fmt.Println("getMsg():\terror on parsing message from <"+userMap[conn].name+"> -> break")
            break
        }

        switch msgPieces[0] {

        // upon REGISTER, we send USERIN to all clients
        case "TCCHAT_REGISTER":
            if len(msgPieces) != 2 || msgPieces[1] == "" {
                fmt.Println("getMsg():\terror on parsing message from <"+userMap[conn].name+"> -> break")
                break
            }
            fmt.Println("getMsg():\tgot 'REGISTER\t"+msgPieces[1]+"' from <"+userMap[conn].name+">")
            mutex.Lock()
            userMap[conn].name = msgPieces[1]
            mutex.Unlock()
            fmt.Println("getMsg():\tuserMap updated -> "+map2string(userMap))
            go sendToAll(userMap, "TCCHAT_USERIN\t"+ userMap[conn].name +"\n")
            fmt.Println("getMsg():\tUSERIN\t"+ userMap[conn].name +" sent to all")


        // upon MESSAGE, we send BCAST to all clients
        case "TCCHAT_MESSAGE":
            if len(msgPieces) != 3 || msgPieces[1] == "" || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                fmt.Println("getMsg():\terror on parsing message (classed MESSAGE) from <"+userMap[conn].name+" -> break")
                break
            }
            fmt.Println("getMsg():\tgot 'MESSAGE\t"+msgPieces[1]+"\t"+msgPieces[2]+"' from <"+userMap[conn].name+">")
            go sendToAll(userMap, "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]+"\n")
            fmt.Println("getMsg():\tBCAST\t"+msgPieces[1]+"\t"+msgPieces[2]+" sent to all")


        // upon PRIVATE, we send PERSONAL to particular client
        case "TCCHAT_PRIVATE":
            if len(msgPieces) != 4 || msgPieces[1] == "" || msgPieces[2] == "" || msgPieces[3] == "" || len(msgPieces[3]) > 140 {
                fmt.Println("getMsg():\terror on parsing message (classed PRIVATE) from <"+userMap[conn].name+" -> break")
                break
            }
            fmt.Println("getMsg():\tgot 'PRIVATE\t"+msgPieces[1]+"\t"+msgPieces[2]+"\t"+msgPieces[3]+"' from <"+userMap[conn].name+">")
            found := false
            wholeMessage = "TCCHAT_PERSONAL\t"+msgPieces[1]+"\t"+msgPieces[3]+"\n"
            for userConn, userClient := range userMap {
                if userClient.name == msgPieces[2] {
                    found = true
                    _ , err := userConn.Write([]byte(wholeMessage))
                    if err != nil {
                        msgCut := strings.TrimSuffix(wholeMessage, "\n")
                        fmt.Println("getMsg():\terror when sending '"+msgCut+"' to <"+msgPieces[2]+"> ("+err.Error()+")")
                    } else {
                        fmt.Println("getMsg():\tPERSONAL\t"+msgPieces[1]+"\t"+msgPieces[3]+" sent to <"+msgPieces[2]+">")
                    }
                    break
                }
            }
            if !found {
                fmt.Println("getMsg():\tcan't deliver pm, user <"+msgPieces[2]+"> not found ! ")
            }


        // upon USERS, we send USERLIST \t client1 \r client2 \r client3 \n
        case "TCCHAT_USERS":
            fmt.Println("getMsg():\tUSERS from <"+userMap[conn].name+">")
            wholeMessage = "TCCHAT_USERLIST\t"+giveUsers(userMap)+"\n"
            _ , err := conn.Write([]byte(wholeMessage))
            if err != nil {
                msgCut := strings.TrimSuffix(wholeMessage, "\n")
                fmt.Println("getMsg():\terror when sending '"+msgCut+"' to <"+userMap[conn].name+"> ("+err.Error()+")")
            } else {                    
                fmt.Println("getMsg():\t"+wholeMessage+" sent to <"+userMap[conn].name+">")
            }


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
            msgCut := strings.TrimSuffix(wholeMessage, "\n")
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

func giveUsers (userMap map[net.Conn]*Client) string {
    var str string
    for _, val := range userMap {
        str += val.name + "\r"
    }
    return strings.TrimSuffix(str, "\r")
}