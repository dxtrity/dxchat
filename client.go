package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/rivo/tview"
)

var (
	app          *tview.Application
	mainTextView *tview.TextView
	inputField   *tview.InputField
	connected    bool
	client       *Client
)

// Client represents a connected client.
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func main() {
	app = tview.NewApplication()

	mainTextView = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetRegions(true)

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorWhite).
		SetPlaceholder("Type your message...")

	grid := tview.NewGrid().
		SetRows(0, 1). // height of each row, in lines
		SetBorders(true)

	// Layout for screens wider than 100 cells.
	grid.AddItem(mainTextView, 0, 0, 1, 1, 0, 100, false).
		AddItem(inputField, 1, 0, 1, 1, 0, 100, true)

	if !connected {
		connectToServer("localhost:12345")
	}

	// Capture key events
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if !connected {
				connectToServer("localhost:12345") // Change server address as needed
			} else {
				if inputField.GetText() == ":quit" {
					client.conn.Close()
					app.Stop()
				}
				sendMessage(inputField.GetText())
				inputField.SetText("")
			}
			return nil // Consume the event
		}
		return event
	})

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}

func connectToServer(serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Printf("Error connecting to server: %v", err)
		return
	}

	client = &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	connected = true

	go receiveMessages()
}

func sendMessage(msg string) {
	if client == nil || client.conn == nil {
		log.Println("Not connected to a server")
		return
	}

	_, err := client.writer.WriteString(msg + "\n")
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}
	client.writer.Flush() // Flush buffer to send the message immediately
}

func receiveMessages() {
	if client == nil || client.conn == nil {
		log.Println("Not connected to a server")
		return
	}

	for {
		msg, err := client.reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if msg == "$BEEP" {
			// MAKE TERMINAL BEEP HERE OR SMTH LIKE THAT
			f, err := os.Open("assets/alert.wav")
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			streamer, format, err := mp3.Decode(f)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			defer streamer.Close()
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			speaker.Play(streamer)
		}

		// Update UI in the application's main goroutine
		app.QueueUpdateDraw(func() {
			currentText := mainTextView.GetText(false) // Get existing text
			newText := currentText + msg               // Append the new message
			mainTextView.SetText(newText)              // Set the updated text

			// Scroll to the end to show the latest messages
			mainTextView.ScrollToEnd()
		})
	}
}
