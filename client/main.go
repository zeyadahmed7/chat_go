
package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"time"

	"go-chat-app/server"
)

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println("Cannot connect to server:", err)
		return
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Could not read username:", err)
		return
	}
	username = strings.TrimSpace(username)

	var welcome string
	err = client.Call("Chat.Join", server.JoinArgs{
		Username: username,
	}, &welcome)

	if err != nil {
		fmt.Println("Join failed:", err)
		return
	}

	fmt.Println(welcome)
	fmt.Println("Commands: /users, /quit")
	fmt.Println("Start chatting!")

	go receiveMessages(client, username)

	for {
		fmt.Print("> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nInput closed.")
			leaveChat(client, username)
			return
		}

		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		switch text {
		case "/users":
			showUsers(client)

		case "/quit":
			leaveChat(client, username)
			fmt.Println("Disconnected from chat.")
			return

		default:
			var reply string
			err := client.Call("Chat.SendMessage", server.SendArgs{
				Username: username,
				Text:     text,
			}, &reply)

			if err != nil {
				fmt.Println("Send failed:", err)
				return
			}
		}
	}
}

func receiveMessages(client *rpc.Client, username string) {
	for {
		var messages []server.Message

		err := client.Call("Chat.Poll", server.UserArgs{
			Username: username,
		}, &messages)

		if err != nil {
			fmt.Println("\nConnection lost:", err)
			return
		}

		for _, message := range messages {
			fmt.Printf("\n[%s] %s\n> ", message.Sender, message.Text)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func showUsers(client *rpc.Client) {
	var users []string

	err := client.Call("Chat.ListUsers", server.Empty{}, &users)
	if err != nil {
		fmt.Println("Could not list users:", err)
		return
	}

	fmt.Println("Connected users:")
	for _, username := range users {
		fmt.Println("-", username)
	}
}

func leaveChat(client *rpc.Client, username string) {
	var reply string

	err := client.Call("Chat.Leave", server.UserArgs{
		Username: username,
	}, &reply)

	if err != nil {
		fmt.Println("Could not leave cleanly:", err)
		return
	}

	fmt.Println(reply)
}
