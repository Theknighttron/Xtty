package client

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"time"
)

// UI represents the terminal UI
type UI struct {
	app           *tview.Application
	messageList   *tview.TextView
	inputField    *tview.InputField
	statusBar     *tview.TextView
	friendList    *tview.List
	user          *User
	currentFriend string
	messages      map[string][]Message
}

// NewUI creates a new terminal UI
func NewUI(user *User) *UI {
	ui := &UI{
		app:      tview.NewApplication(),
		user:     user,
		messages: make(map[string][]Message),
	}
	ui.messageList = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			ui.app.QueueUpdateDraw(func() {})
		})

	ui.inputField = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)

	ui.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[yellow]Xtty - Room Code: [white]%s | [yellow]Status: [white]Waiting for peer...", user.RoomCode))

	ui.friendList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedBackgroundColor(tcell.ColorNavy)

	// Start listening for incoming messages
	go ui.updateMessages()

	return ui
}

// setupUI sets up the terminal UI
func (ui *UI) setupUI() {
	// Setup message list
	ui.messageList.SetBorder(true).SetTitle("Messages")

	// Setup input field
	ui.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := ui.inputField.GetText()
			if text == "" {
				return
			}

			if strings.HasPrefix(text, "/") {
				ui.handleCommand(text)
			} else {
				if ui.currentFriend == "" && ui.friendList.GetItemCount() > 0 {
					// Auto-select the first friend if none is selected
					mainText, _ := ui.friendList.GetItemText(0)
					ui.currentFriend = mainText
					ui.friendList.SetCurrentItem(0)
					ui.statusBar.SetText(fmt.Sprintf("[yellow]Xtty - Chatting with %s[white]", ui.currentFriend))
				}

				if ui.currentFriend == "" {
					ui.messageList.Write([]byte("[red]Error: No connection established yet[white]\n"))
					ui.inputField.SetText("")
					return
				}

				// Send message
				if err := ui.user.SendMessage(text); err != nil {
					ui.messageList.Write([]byte(fmt.Sprintf("[red]Error sending message: %v[white]\n", err)))
				} else {
					ui.displayMessage(text, true)
				}
				ui.inputField.SetText("")
			}
		}
	})

	// Setup friend list
	ui.friendList.SetBorder(true).SetTitle("Connections")

	// Add connected peer to the list when connection is established
	if ui.user.PeerPubKey != nil {
		peerName := "Peer"
		ui.friendList.AddItem(peerName, "", 0, func() {
			ui.selectFriend(peerName)
		})
	}

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.statusBar, 1, 1, false).
		AddItem(tview.NewFlex().
			AddItem(ui.friendList, 20, 1, true).
			AddItem(ui.messageList, 0, 3, false),
			0, 1, false).
		AddItem(ui.inputField, 1, 1, true)

	// Set the root and focus
	ui.app.SetRoot(flex, true)
	ui.app.SetFocus(ui.inputField)

	// Set key bindings
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// Toggle focus between friend list and input field
			if ui.app.GetFocus() == ui.friendList {
				ui.app.SetFocus(ui.inputField)
			} else {
				ui.app.SetFocus(ui.friendList)
			}
			return nil
		case tcell.KeyCtrlQ:
			// Quit the application
			ui.app.Stop()
			return nil
		}
		return event
	})
}

// Run starts the UI
func (ui *UI) Run() error {
	// Setup UI
	ui.setupUI()

	// Start the application
	return ui.app.Run()
}

// selectFriend selects a friend to chat with
func (ui *UI) selectFriend(friendName string) {
	ui.currentFriend = friendName
	ui.statusBar.SetText(fmt.Sprintf("[yellow]Xtty - Connected with %s[white]", friendName))
	ui.messageList.Clear()

	// Display messages for this friend
	if messages, ok := ui.messages[friendName]; ok {
		for _, msg := range messages {
			ui.displayLocalMessage(&msg)
		}
	} else {
		// Also display messages from the User's message list
		for _, msg := range ui.user.Messages {
			ui.displayMessage(msg.Content, msg.Sent)
		}
	}

	ui.app.SetFocus(ui.inputField)
}

// updateMessages monitors for new messages from the User
func (ui *UI) updateMessages() {
	for range ui.user.Done {
		ui.app.QueueUpdateDraw(func() {
			// Get the latest message
			if len(ui.user.Messages) > 0 {
				latestMsg := ui.user.Messages[len(ui.user.Messages)-1]
				ui.displayMessage(latestMsg.Content, latestMsg.Sent)

				// Update the friend list if this is a new connection
				if ui.friendList.GetItemCount() == 0 && ui.user.PeerPubKey != nil {
					peerName := "Peer"
					ui.friendList.AddItem(peerName, "", 0, func() {
						ui.selectFriend(peerName)
					})

					// Auto-select the first connection
					ui.selectFriend(peerName)
				}
			}
		})
	}
}

// displayMessage displays a message in the UI
func (ui *UI) displayMessage(content string, sent bool) {
	// Format message based on sender
	var formattedMessage string
	timestamp := time.Now().Format("15:04:05")

	if sent {
		formattedMessage = fmt.Sprintf("[blue]%s [You][white]: %s\n", timestamp, content)
	} else {
		formattedMessage = fmt.Sprintf("[green]%s [Peer][white]: %s\n", timestamp, content)
	}

	// Write to message list
	ui.messageList.Write([]byte(formattedMessage))

	// Store message in local format
	msg := Message{
		Content:   content,
		Timestamp: time.Now(),
		Sent:      sent,
	}

	// Store message in our local map
	if ui.currentFriend != "" {
		if _, ok := ui.messages[ui.currentFriend]; !ok {
			ui.messages[ui.currentFriend] = make([]Message, 0)
		}
		ui.messages[ui.currentFriend] = append(ui.messages[ui.currentFriend], msg)
	}
}

// displayLocalMessage displays a previously stored message
func (ui *UI) displayLocalMessage(message *Message) {
	// Format timestamp
	timestamp := message.Timestamp.Format("15:04:05")

	// Format message based on sender
	var formattedMessage string
	if message.Sent {
		formattedMessage = fmt.Sprintf("[blue]%s [You][white]: %s\n", timestamp, message.Content)
	} else {
		formattedMessage = fmt.Sprintf("[green]%s [Peer][white]: %s\n", timestamp, message.Content)
	}

	// Write to message list
	ui.messageList.Write([]byte(formattedMessage))
}

// updateStatus updates the status bar text
func (ui *UI) updateStatus(text string) {
	ui.app.QueueUpdateDraw(func() {
		ui.statusBar.SetText(text)
	})
}

// handleCommand processes command inputs
func (ui *UI) handleCommand(cmd string) {
	switch {
	case cmd == "/quit":
		ui.user.Cleanup()
		ui.app.Stop()
	case cmd == "/help":
		ui.displayMessage("Available commands: /quit, /help, /clear", true)
	case cmd == "/clear":
		ui.messageList.Clear()
	case strings.HasPrefix(cmd, "/connect"):
		parts := strings.SplitN(cmd, " ", 2)
		if len(parts) < 2 {
			ui.displayMessage("Usage: /connect [SERVER_URL] [ROOM_CODE]", true)
		} else {
			// Implementation would depend on your connection logic
			ui.displayMessage("Connection command received (not implemented yet)", true)
		}
	default:
		ui.displayMessage(fmt.Sprintf("Unknown command: %s", cmd), true)
	}
}

// Stop stops the UI
func (ui *UI) Stop() {
	ui.app.Stop()
}
