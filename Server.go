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
    registered bool
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
        if err != nil {
            fmt.Println("getConn():\terror when accepting new connection -> continue\n"+err.Error())
            continue
        } else {
            fmt.Println("getConn():\tnew connection "+conn.RemoteAddr().String())
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
            fmt.Println("getConn():\terror when sending TCCHAT_WELCOME to <"+userMap[conn].name+"> -> continue\n"+firstMessageErr.Error())
            continue
        }

        // 3) we listen to it, because it's supposed to send REGISTER right after connecting
        go getMsg(serverName, conn, userMap,mutex)
    }
}


func getMsg(serverName *string, conn net.Conn, userMap map[net.Conn]*Client, mutex *sync.Mutex) {
    var userName string = userMap[conn].name
    var msgPieces []string
    var wholeMessage string
    var alreadyRegistered bool
    reader := bufio.NewReader(conn)

    fmt.Println("getMsg():\tlistening on messages from <"+userName+">")
    for {
        // receiving and parsing the new message
        text, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("getMsg():\terror on reading message from <"+userName+"> ("+err.Error()+") -> force disconnect")
            go sendToAll(userMap, mutex, "TCCHAT_USEROUT\t"+ userName +"\n")
            disconnect(conn, userName, userMap, mutex)
            return
        }

        text = strings.TrimSuffix(text, "\n")
        msgPieces = strings.SplitN(text, "\t", 4)
        if len(msgPieces) < 1 || msgPieces[0] == "" {
            fmt.Println("getMsg():\terror on parsing message from <"+userName+"> -> continue")
            continue
        }

        switch msgPieces[0] {

        // upon REGISTER, we send USERIN to all clients
        case "TCCHAT_REGISTER":
            if len(msgPieces) != 2 || msgPieces[1] == "" {
                fmt.Println("getMsg():\terror on parsing message (classed REGISTER) from <"+userName+"> -> continue")
                continue
            } else {
                fmt.Println("\033[32mgetMsg():\033[0m\tgot "+text+" from <"+userName+">")
            }

            mutex.Lock()
            userMap[conn].name = msgPieces[1]
            userName = msgPieces[1]
            alreadyRegistered = userMap[conn].registered
            userMap[conn].registered = true
            mutex.Unlock()
            fmt.Println("\033[32mgetMsg():\033[0m\tuserMap updated -> "+map2string(userMap))

            if alreadyRegistered {
                go sendToAll(userMap, mutex, "TCCHAT_USEROUT\t"+ userName +"\n")
                disconnect(conn, userName, userMap, mutex)
                return
            }
            go sendToAll(userMap, mutex, "TCCHAT_USERIN\t"+ userName +"\n")


        // upon MESSAGE, we send BCAST to all clients
        case "TCCHAT_MESSAGE":
            if len(msgPieces) != 3 || msgPieces[1] == "" || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                fmt.Println("getMsg():\terror on parsing message (classed MESSAGE) from <"+userName+" -> continue")
                continue
            } else {
                fmt.Println("getMsg():\tgot "+text+" from <"+userName+">")
            }

            go sendToAll(userMap, mutex, "TCCHAT_BCAST\t"+msgPieces[1]+"\t"+msgPieces[2]+"\n")


        // upon BAN, we send USERBAN to all clients and disconnect the user in question
        case "TCCHAT_BAN":
            if len(msgPieces) != 3 || msgPieces[1] == "" || msgPieces[2] == "" {
                fmt.Println("getMsg():\terror on parsing message (classed BAN) from <"+userName+" -> continue")
                continue
            } else {
                fmt.Println("getMsg():\tgot "+text+" from <"+userName+">")
            }

            destConn, destName, found := findUser(msgPieces[2], userMap, mutex)
            if !found {
                fmt.Println("getMsg():\tcan't ban, user <"+msgPieces[2]+"> not found ! ")
            } else {
                go sendToAll(userMap, mutex, "TCCHAT_USERBAN\t"+msgPieces[1]+"\t"+msgPieces[2]+"\n")
                disconnect(destConn, destName, userMap, mutex)
            }


        // upon PRIVATE, we send PERSONAL to particular client
        case "TCCHAT_PRIVATE":
            if len(msgPieces) != 4 || msgPieces[1] == "" || msgPieces[2] == "" || msgPieces[3] == "" || len(msgPieces[3]) > 140 {
                fmt.Println("getMsg():\terror on parsing message (classed PRIVATE) from <"+userName+" -> continue")
                continue
            } else {
                fmt.Println("getMsg():\tgot "+text+" from <"+userName+">")
            }

            destConn, destName, found := findUser(msgPieces[2], userMap, mutex)
            if !found {
                fmt.Println("getMsg():\tcan't deliver pm, user <"+msgPieces[2]+"> not found ! ")
            } else {
                sendTo(destConn, destName, "TCCHAT_PERSONAL\t"+msgPieces[1]+"\t"+msgPieces[3]+"\n")
            }


        // upon USERS, we send TCCHAT_USERLIST \t client1 \r client2 \r client3 \n
        case "TCCHAT_USERS":
            fmt.Println("getMsg():\tgot USERS from <"+userName+">")
            wholeMessage = "TCCHAT_USERLIST\t"+giveUsers(userMap)+"\n"
            sendTo(conn, userName, wholeMessage)


        // upon DISCONNECT, send USEROUT to all clients and disconnect it
        case "TCCHAT_DISCONNECT":
            fmt.Println("\033[31mgetMsg():\033[0m\t\tgot DISCONNECT from <"+userName+">")
            go sendToAll(userMap, mutex, "TCCHAT_USEROUT\t"+ userName +"\n")
            disconnect(conn, userName, userMap, mutex)
            return
        }
    }
}

func sendTo(conn net.Conn, userName string, wholeMessage string) {
    messageCut := strings.TrimSuffix(wholeMessage, "\n")

    _ , err := conn.Write([]byte(wholeMessage))
    if err != nil {
        fmt.Println("sendTo():\terror when sending '"+messageCut+"' to <"+userName+"> ("+err.Error()+")")
    } else {
        // si on affiche la liste des utilisateurs, il faut remplacer les \r
        messageCut = strings.Replace(messageCut,"\r",", ",-1)
        fmt.Println("sendTo():\t'"+messageCut+"' successfully sent to <"+userName+">")
    }
}

func sendToAll(userMap map[net.Conn]*Client, mutex *sync.Mutex, wholeMessage string) {
    ok := true
    messageCut := strings.TrimSuffix(wholeMessage, "\n")

    mutex.Lock()
    for userConn, userClient := range userMap {
        _ , err := userConn.Write([]byte(wholeMessage))
        if err != nil {
            ok = false
            fmt.Println("sendToAll():\terror when sending '"+messageCut+"' to <"+userClient.name+"> ("+err.Error()+")")
        }
    }
    mutex.Unlock()
    if ok {
        fmt.Println("sendToAll():\t'"+messageCut+"' successfully sent to all users.")
    }
}

func disconnect(conn net.Conn, userName string, userMap map[net.Conn]*Client, mutex *sync.Mutex) {
    mutex.Lock()
    _, ok := userMap[conn]
    mutex.Unlock()
    if ok {
        mutex.Lock()
        delete(userMap, conn)
        mutex.Unlock()
        fmt.Println("\033[31mdisconnect():\033[0m\tsuccessfully disconnected <"+userName+">")
    } else {
        fmt.Println("disconnect():\terror on disconnecting <"+userName+">")
        fmt.Println("disconnect():\tuserMap is now "+map2string(userMap))
    }
}

func findUser(userName string, userMap map[net.Conn]*Client, mutex *sync.Mutex) (net.Conn, string, bool) {
    mutex.Lock()

    for userConn, userClient := range userMap {
        if userClient.name == userName {
            mutex.Unlock()
            return userConn, userName, true
        }
    }

    mutex.Unlock()
    return nil, "", false
}

func giveUsers (userMap map[net.Conn]*Client) string {
    var str string
    for _, val := range userMap {
        str += val.name + "\r"
    }
    return strings.TrimSuffix(str, "\r")
}

func map2string(userMap map[net.Conn]*Client) string {
    var str string = "[ "
    for key, val := range userMap {
        str += key.RemoteAddr().String()+":"+val.name+", "
    }
    return str + " ]"
}
