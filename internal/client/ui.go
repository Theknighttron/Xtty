package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app         *tview.Application
	messageView *tview.TextView
	inputField  *tview.InputField
	statusView  *tview.TextView
	user        *User
}

func NewUI(user *User) *UI {
	ui := &UI{
		app:  tview.NewApplication(),
		user: user,
	}

	ui.messageView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)

	ui.inputField = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)

	ui.statusView = tview.NewTextView().
		SetDynamicColors(true)

	return ui
}

func (ui *UI) Run() error {
	ui.updateStatus()

	// Wait for initial connection
	if ui.user.PeerPubKey == nil && ui.user.RoomCode != "" {
		ui.displaySystemMessage("Establishing secure connection...")
		select {
		case <-ui.user.KeyExchangeDone:
			ui.displaySystemMessage("Secure connection established!")
		case <-time.After(10 * time.Second):
			ui.displaySystemMessage("Warning: Connection not fully secured")
		}
	}

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.statusView, 1, 1, false).
		AddItem(ui.messageView, 0, 1, false).
		AddItem(ui.inputField, 1, 1, true)

	ui.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := ui.inputField.GetText()
			if text == "" {
				return
			}

			if strings.HasPrefix(text, "/") {
				ui.handleCommand(text)
			} else {
				if err := ui.user.SendMessage(text); err != nil {
					ui.displaySystemMessage(fmt.Sprintf("Error: %v", err))
				}
			}
			ui.inputField.SetText("")
		}
	})

	go ui.messagePoller()

	ui.app.SetRoot(flex, true)
	return ui.app.Run()
}

func (ui *UI) handleCommand(cmd string) {
	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "/join":
		if len(parts) < 2 {
			ui.displaySystemMessage("Usage: /join ROOM_CODE")
			return
		}
		ui.displaySystemMessage(fmt.Sprintf("Joining room: %s", parts[1]))
		if err := ui.user.JoinRoom("ws://localhost:8080", parts[1]); err != nil {
			ui.displaySystemMessage(fmt.Sprintf("Join failed: %v", err))
		} else {
			// Wait for key exchange after joining
			go func() {
				select {
				case <-ui.user.KeyExchangeDone:
					ui.app.QueueUpdateDraw(func() {
						ui.displaySystemMessage("Successfully joined room")
						ui.updateStatus()
					})
				case <-time.After(10 * time.Second):
					ui.app.QueueUpdateDraw(func() {
						ui.displaySystemMessage("Warning: Secure connection not established")
					})
				}
			}()
		}
	case "/help":
		ui.displaySystemMessage("Commands:\n/join ROOM_CODE - Join a room\n/help - Show this help")
	default:
		ui.displaySystemMessage(fmt.Sprintf("Unknown command: %s", parts[0]))
	}
}

func (ui *UI) messagePoller() {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ui.user.Done:
			return
		case <-ticker.C:
			ui.app.QueueUpdateDraw(func() {
				for _, msg := range ui.user.Messages {
					ui.displayMessage(msg)
				}
				ui.user.Messages = nil // Clear displayed messages
				ui.updateStatus()
			})
		}
	}
}

func (ui *UI) displayMessage(msg Message) {
	var prefix, color string
	if msg.Sent {
		color = "[blue]"
		prefix = ui.user.Username
	} else {
		color = "[green]"
		prefix = msg.Sender
	}
	fmt.Fprintf(ui.messageView, "%s%s[white]: %s\n", color, prefix, msg.Content)
}

func (ui *UI) displaySystemMessage(text string) {
	fmt.Fprintf(ui.messageView, "[yellow]SYSTEM[white]: %s\n", text)
}

func (ui *UI) updateStatus() {
	status := fmt.Sprintf("[yellow]%s[white] | Room: %s", ui.user.Username, ui.user.RoomCode)
	if ui.user.PeerPubKey != nil {
		status += " | [green]Connected[white]"
	} else if ui.user.RoomCode != "" {
		status += " | [yellow]Waiting for peer...[white]"
	} else {
		status += " | [red]Disconnected[white]"
	}
	ui.statusView.SetText(status)
}
