package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Client represents a connected client.
type Client struct {
	conn     net.Conn
	nickname string
	color    string
	recv     chan string
}

var (
	clients       = make(map[net.Conn]*Client)
	colors        = []string{"red", "green", "aqua", "yellow", "cyan", "pink", "lime", "purple"}
	clientsLock   sync.Mutex
	messagesQueue = make(chan string, 100)
)

func main() {
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	fmt.Println("SERVER STARTED ON localhost:12345")

	go handleMessagesBackup()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleClient(conn)
	}
}

/* Checks the name of the nickname with regex to ensure it is valid */
func checkNickname(str string) bool {
	regex := regexp.MustCompile(`[\w]+`)
	return regex.MatchString(str)
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	if len(clients) >= 8 {
		io.WriteString(conn, "Server is full. Try again later.\n")
		return
	}

	fmt.Printf("Connection made.\n")

	client := &Client{
		conn: conn,
		recv: make(chan string),
	}

	clientsLock.Lock()
	clients[conn] = client
	clientsLock.Unlock()

	// Assign a unique color
	client.color = assignColor()

	// Ask client for nickname
	io.WriteString(conn, "Welcome! Please choose a nickname: \n")

Nickname:
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		nickname := scanner.Text()
		if strings.TrimSpace(nickname) != "" && checkNickname(nickname) {
			client.nickname = nickname
			break
		}
		io.WriteString(conn, "[red]Invalid nickname. Please choose another one.[white] \n")
		goto Nickname
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading nickname: %v", err)
		return
	}

	io.WriteString(conn, fmt.Sprintf("Your nickname is [%s]%s[white] and your color is [%s]%s[white].\n", client.color, client.nickname, client.color, client.color))

	fmt.Printf("Client: %s connected at %s with nickname %s\n", conn.LocalAddr(), time.Now().Format("15:04:05"), client.nickname)

	// Listen for incoming messages
	go func() {
		for msg := range client.recv {
			// Broadcast message to all clients
			broadcastMessage(fmt.Sprintf("[%s] [%s]%s[white]: %s", time.Now().Format("15:04"), client.color, client.nickname, msg))
		}
	}()

	// Read messages from client
	scanner = bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		client.recv <- msg
	}

	clientsLock.Lock()
	delete(clients, conn)
	clientsLock.Unlock()
}

func broadcastMessage(msg string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for _, client := range clients {
		go func(c *Client) {
			io.WriteString(c.conn, msg+"\n")
		}(client)
	}

	// Queue message for backup
	messagesQueue <- msg
}

func assignColor() string {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for _, c := range colors {
		colorTaken := false
		for _, client := range clients {
			if client.color == c {
				colorTaken = true
				break
			}
		}
		if !colorTaken {
			return c
		}
	}
	return "gray" // Default color if all colors are taken
}

func handleMessagesBackup() {
	backupInterval := 10 * time.Minute // Adjust as needed (e.g., 3 minutes)
	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()

	var messages []string

	for {
		select {
		case msg := <-messagesQueue:
			messages = append(messages, msg)
		case <-ticker.C:
			if len(messages) > 0 {
				writeMessagesToBackup(messages)
				messages = nil // Clear the messages slice after writing to backup
			}
		}
	}
}

func writeMessagesToBackup(messages []string) {
	fileName := fmt.Sprintf("messages_%s.txt", time.Now().Format("20060102_150405"))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file for backup: %v", err)
		return
	}
	defer file.Close()

	for _, msg := range messages {
		if _, err := file.WriteString(msg + "\n"); err != nil {
			log.Printf("Error writing to backup file: %v", err)
			// Continue writing other messages even if one fails
		}
	}

	fmt.Printf("Chat backup made at: %s\n", time.Now().Format("15:04:05"))
}
