package client

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Theknighttron/Xtty/internal/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI represents the terminal UI
type UI struct {
	app           *tview.Application
	messageList   *tview.TextView
	inputField    *tview.InputField
	statusBar     *tview.TextView
	friendList    *tview.List
	client        *Client
	currentFriend string
	messages      map[string][]common.Message
}

// NewUI creates a new terminal UI
func NewUI(client *Client) *UI {
	ui := &UI{
		app:      tview.NewApplication(),
		client:   client,
		messages: make(map[string][]common.Message),
		messageList: tview.NewTextView().
			SetDynamicColors(true).
			SetChangedFunc(func() {
				ui.app.Draw()
			}),
		inputField: tview.NewInputField().
			SetLabel("> ").
			SetFieldWidth(0),
		statusBar: tview.NewTextView().
			SetDynamicColors(true).
			SetText("[yellow]Xtty - Secure Terminal Chat[white]"),
		friendList: tview.NewList().
			ShowSecondaryText(false).
			SetHighlightFullLine(true).
			SetSelectedBackgroundColor(tcell.ColorNavy),
	}

	// Configure message handler
	client.messageHandler = ui.handleIncomingMessage

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

			if ui.currentFriend == "" {
				ui.messageList.Write([]byte("[red]Error: No friend selected[white]\n"))
				ui.inputField.SetText("")
				return
			}

			// Send message
			ui.sendMessage(text)
			ui.inputField.SetText("")
		}
	})

	// Setup friend list with some dummy friends
	ui.friendList.SetBorder(true).SetTitle("Friends")

	// Add some dummy friends for testing
	ui.friendList.AddItem("alice", "", 0, func() {
		ui.selectFriend("alice")
	})
	ui.friendList.AddItem("bob", "", 0, func() {
		ui.selectFriend("bob")
	})
	ui.friendList.AddItem("charlie", "", 0, func() {
		ui.selectFriend("charlie")
	})

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
	ui.app.SetFocus(ui.friendList)

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
	ui.statusBar.SetText(fmt.Sprintf("[yellow]Xtty - Chatting with %s[white]", friendName))
	ui.messageList.Clear()

	// Display messages for this friend
	if messages, ok := ui.messages[friendName]; ok {
		for _, msg := range messages {
			ui.displayMessage(&msg)
		}
	}

	ui.app.SetFocus(ui.inputField)
}

// sendMessage sends a message to the current friend
func (ui *UI) sendMessage(text string) {
	if ui.currentFriend == "" {
		return
	}

	// Create and send message
	if err := ui.client.SendMessage(ui.currentFriend, common.TypeText, text); err != nil {
		ui.messageList.Write([]byte(fmt.Sprintf("[red]Error sending message: %v[white]\n", err)))
		return
	}

	// Create local message
	message := common.Message{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		SenderID:    ui.client.config.Username,
		RecipientID: ui.currentFriend,
		Type:        common.TypeText,
		Timestamp:   time.Now(),
		Content:     text,
	}

	// Store and display message
	ui.storeMessage(&message)
	ui.displayMessage(&message)
}

// handleIncomingMessage handles incoming messages
func (ui *UI) handleIncomingMessage(message *common.Message) {
	ui.app.QueueUpdateDraw(func() {
		// Store message
		ui.storeMessage(message)

		// If this message is from/to the current friend, display it
		if ui.currentFriend == message.SenderID || ui.currentFriend == message.RecipientID {
			ui.displayMessage(message)
		}

		// Update friend list to show unread messages
		// This would be implemented in a real app
	})
}

// storeMessage stores a message
func (ui *UI) storeMessage(message *common.Message) {
	// Determine the key to use for this message
	key := message.SenderID
	if message.SenderID == ui.client.config.Username {
		key = message.RecipientID
	}

	// Ensure we have a slice for this friend
	if _, ok := ui.messages[key]; !ok {
		ui.messages[key] = make([]common.Message, 0)
	}

	// Store the message
	ui.messages[key] = append(ui.messages[key], *message)
}

// displayMessage displays a message in the UI
func (ui *UI) displayMessage(message *common.Message) {
	// Format timestamp
	timestamp := message.Timestamp.Format("15:04:05")

	// Format message based on sender
	var formattedMessage string
	if message.SenderID == ui.client.config.Username {
		formattedMessage = fmt.Sprintf("[blue]%s [You][white]: %s\n", timestamp, message.Content)
	} else {
		formattedMessage = fmt.Sprintf("[green]%s [%s][white]: %s\n", timestamp, message.SenderID, message.Content)
	}

	// Write to message list
	ui.messageList.Write([]byte(formattedMessage))
}

// Stop stops the UI
func (ui *UI) Stop() {
	ui.app.Stop()
}
