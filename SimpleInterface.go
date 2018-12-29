package main

import (
	//"fmt"
	"log"
	"time"
    "strconv"

	"github.com/marcusolsson/tui-go"
)

type post struct {
	username string
	message  string
	time     string
}

var(
	input = tui.NewEntry()
	history = tui.NewVBox()
	messages = make(chan string)

	posts = []post{
		{username: "john", message: "hi, what's up?", time: "14:41"},
		{username: "jane", message: "not much", time: "14:43"},
	}
)

func main() {
	sidebar := tui.NewVBox(
		tui.NewLabel("SERVER\nTc-chatcommunity"),
		tui.NewLabel("TC-Chat community server"),
		tui.NewLabel(""),
		tui.NewLabel("USERS"),
		tui.NewLabel("slackbot"),
        tui.NewSpacer(),
	)
	sidebar.SetBorder(true)


	for _, m := range posts {
		history.Append(tui.NewLabel(m.time + " <" + m.username + "> "+ m.message))
	}

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)


	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)


	input.OnSubmit(func(e *tui.Entry) {
		history.Append(tui.NewLabel(time.Now().Format("15:04") + " <john> "+ e.Text()))
		input.SetText("")
	})


	root := tui.NewHBox(sidebar, chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

    go bonus()


	go func(){
        for{
		select{
		case msg := <-messages:
			history.Append(tui.NewLabel(time.Now().Format("15:04") + " <john> "+ msg))
            ui.Repaint()
		}
	    }
    }()

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

}

func bonus(){
    n := 12
    str := ""
	for{
		time.Sleep(1000 * time.Millisecond)
        str = "test "+strconv.Itoa(n)
		 messages <- str
         n = n+1

	}
}

