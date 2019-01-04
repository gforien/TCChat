package main

import (
    "strings"
    "fmt"
    "flag"
    "net"
    "bufio"
    "log"
    "time"
    "strconv"
    "os"
    "math/rand"

    "github.com/marcusolsson/tui-go"
)

func main() {
    var (
        nickname *string            // the user name
        conn net.Conn               // the socket, we send messages via conn.Write()
        connectionErr error
        history *tui.Box            // the main window, we add new messages to this window via history.Append()
        serverName *tui.Label       // the Label in the sidebar containing the server name
        userList *tui.Label         // the Label in the sidebar containing the user list
        input *tui.Entry            // the bottom bar, we treat user input via input.onSubmit()
    )

    //------------------------------------------------//
    //           1. Graphical initialization
    //------------------------------------------------//
    serverName = tui.NewLabel("undefined")
    userList = tui.NewLabel("undefined")
    sidebar := tui.NewVBox(
        tui.NewLabel("SERVER"),
        serverName,
        tui.NewLabel(""),
        tui.NewLabel("USERS"),
        userList,
        tui.NewSpacer(),
    )
    sidebar.SetBorder(true)

    history = tui.NewVBox()
    historyScroll := tui.NewScrollArea(history)
    historyScroll.SetAutoscrollToBottom(true)
    historyBox := tui.NewVBox(historyScroll)
    historyBox.SetBorder(true)

    input = tui.NewEntry()
    input.SetFocused(true)
    input.SetSizePolicy(tui.Expanding, tui.Maximum)
    inputBox := tui.NewHBox(input)
    inputBox.SetBorder(true)
    inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

    // pack the whole thing in a Box and create ui object
    chat := tui.NewVBox(historyBox, inputBox)
    chat.SetSizePolicy(tui.Expanding, tui.Expanding)
    root := tui.NewHBox(sidebar, chat)
    ui, uiErr := tui.New(root)
    if uiErr != nil {
        log.Fatal(uiErr)
    }

    ui.SetKeybinding("Esc", func() { ui.Quit() })


    //------------------------------------------------//
    //             2. Client initialization
    //------------------------------------------------//
    // if the client doesn't provide a nickname, we define a random one
    rand.Seed(time.Now().Unix())
    randomNickname := "client-"+strconv.Itoa(rand.Intn(9999))

    // server adress and nickname are command-line arguments
    address := flag.String("address", "127.0.0.1:2000", "IP address and port of the server")
    nickname = flag.String("nickname", randomNickname, "nickname used to identify yourself")
    flag.Parse()

    conn, connectionErr = net.Dial("tcp", *address)
    if connectionErr != nil {
        fmt.Println("Essayez de lancer le serveur d'abord ;)")
        panic(connectionErr)
    }

    // messages are received and treated in an infinite loop
    go getMsg(conn, history, serverName, userList, ui)

    // when <Enter> is pressed, text is treated by getInput() and input buffer is flushed
    input.OnSubmit(func(e *tui.Entry) {
        getInput(e.Text(), nickname, conn, history)
        input.SetText("")
    })

    // we have to send a first message to register the nickname to the server
    _ , firstMessageErr := conn.Write([]byte("TCCHAT_REGISTER\t"+ *nickname +"\n"))
    if firstMessageErr != nil {
        fmt.Println("Error in main(), to register username\n"+ firstMessageErr.Error())
    }
    // fetch user list
    conn.Write([]byte("TCCHAT_USERS\n"))

    // launch the graphical mainloop()
    if mainloopErr := ui.Run(); mainloopErr != nil {
        log.Fatal(mainloopErr)
    }
}


func getInput(text string, nickname *string, conn net.Conn, history *tui.Box) {
    var msgPieces []string
    var err error
    help := `/help : print this help page
            /disconnect : close the client
            /users : print the list of connected users
            /mp <recipient> <TAB> <message_payload> : send a private message to a given recipient`

    // user is not allowed to send an empty string
    if text == "" {
        return
    }

    // if user typed a command, we treat it accordingly
    if strings.HasPrefix(text,"/") {
        msgPieces = strings.SplitN(text, " ", 2)
        switch msgPieces[0] {

        case "/help" :
            history.Append(tui.NewLabel(help))

        case "/ban" :
            _, err = conn.Write([]byte("TCCHAT_BAN\t"+*nickname+"\t"+msgPieces[1]+"\n"))
            if err != nil {
                fmt.Println("Error in getInput() case /ban\n"+err.Error())
                return
            }

        case "/disconnect" :
            _, err = conn.Write([]byte("TCCHAT_DISCONNECT\n"))
            if err != nil {
                fmt.Println("Error in getInput() case /disconnect\n")
                panic(err)
            }
            os.Exit(0)

        case "/mp" :
            msgPieces = strings.SplitN(msgPieces[1], " ", 2)
            _, err = conn.Write([]byte("TCCHAT_PRIVATE\t"+*nickname+"\t"+msgPieces[0]+"\t"+msgPieces[1]+"\n"))
            if err != nil {
                fmt.Println("Error in getInput() case /mp\n"+err.Error())
                return
            }
            history.Append(tui.NewLabel("to "+msgPieces[0]+" (in private) : "+msgPieces[1]))

        case "/users" :
            _, err = conn.Write([]byte("TCCHAT_USERS\n"))
            if err != nil {
                fmt.Println("Error in getInput() case /users\n"+err.Error())
                return
            }

        case "/raw" :
            _, err = conn.Write([]byte(msgPieces[1]+"\n"))
            if err != nil {
                fmt.Println("Error in getInput() case /raw\n"+err.Error())
                return
            }

        default :
            history.Append(tui.NewLabel("Undefined command. Try /help to see available ones."))
        }
    } else {
    // if it's not command, we just check the size and send the whole text as a TCCHAT_MESSAGE
        if len(text) > 140 {
            history.Append(tui.NewLabel("Error : your message could not be sent because it has more than 140 characters."))
            return
        }
        _, err = conn.Write([]byte("TCCHAT_MESSAGE\t"+*nickname+"\t"+text+"\n"))
        if err != nil {
            fmt.Println("Error in getInput(), block else{}\n"+err.Error())
            return
        }
    }
}


func getMsg(conn net.Conn, history *tui.Box, serverName *tui.Label, userList *tui.Label, ui tui.UI) {
    var msgPieces []string
    reader := bufio.NewReader(conn)
    invalidProtocol := "Received message doesn't respect TC-Chat protocol."

    for {
        text, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error in getMsg(), at ReadString()\n"+err.Error())
            return
        }
        text = strings.TrimSuffix(text, "\n")
        msgPieces = strings.SplitN(text, "\t", 3)

        if len(msgPieces) < 2 || msgPieces[1] == "" {msgPieces = make([]string, 1)}

        switch msgPieces[0] {
            case "TCCHAT_WELCOME":
                history.Append(tui.NewLabel("Welcome on the server : " + msgPieces[1]))
                serverName.SetText(msgPieces[1])

            case "TCCHAT_USERIN":
                history.Append(tui.NewLabel("User in : " + strings.Split(msgPieces[1], "\n")[0]))
                conn.Write([]byte("TCCHAT_USERS\n"))

            case "TCCHAT_USEROUT":
                history.Append(tui.NewLabel("User out : " + msgPieces [1]))
                conn.Write([]byte("TCCHAT_USERS\n"))

            case "TCCHAT_USERBAN":
                history.Append(tui.NewLabel("User out : " + msgPieces [2] + "(banned by " + msgPieces[1] + ")"))
                conn.Write([]byte("TCCHAT_USERS\n"))

            case "TCCHAT_USERLIST":
                userList.SetText(strings.Replace(msgPieces[1],"\r","\n",-1))

            case "TCCHAT_BCAST":
                if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                    fmt.Println(invalidProtocol)
                    continue
                } else {
                    history.Append(tui.NewLabel(msgPieces[1]+" says : "+msgPieces[2]))
                }

            case "TCCHAT_PERSONAL" :
                if len(msgPieces) != 3 || msgPieces[2] == "" || len(msgPieces[2]) > 140 {
                    fmt.Println(invalidProtocol)
                    continue
                } else {
                    history.Append(tui.NewLabel(msgPieces[1]+" says (in private) : "+msgPieces[2]))
                }

            default :
                fmt.Println(invalidProtocol)
                continue
        }
        ui.Repaint()
    }
}
