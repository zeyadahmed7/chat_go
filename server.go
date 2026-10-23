
package server

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Sender string
	Text   string
}

type JoinArgs struct {
	Username string
}

type UserArgs struct {
	Username string
}

type SendArgs struct {
	Username string
	Text     string
}

type Empty struct{}

type ClientState struct {
	Queue    []Message
	LastSeen time.Time
}

type ChatService struct {
	mu      sync.Mutex
	clients map[string]*ClientState
}

func NewChatService() *ChatService {
	return &ChatService{
		clients: make(map[string]*ClientState),
	}
}

func (s *ChatService) Join(args JoinArgs, reply *string) error {
	username := strings.TrimSpace(args.Username)

	if username == "" {
		return errors.New("username cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[username]; exists {
		return errors.New("username already exists")
	}

	s.clients[username] = &ClientState{
		LastSeen: time.Now(),
	}

	for name, client := range s.clients {
		if name != username {
			client.Queue = append(client.Queue, Message{
				Sender: "SERVER",
				Text:   username + " joined the chat",
			})
		}
	}

	*reply = "Welcome, " + username
	return nil
}

func (s *ChatService) SendMessage(args SendArgs, reply *string) error {
	text := strings.TrimSpace(args.Text)

	if text == "" {
		return errors.New("message cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sender, exists := s.clients[args.Username]
	if !exists {
		return errors.New("user is not connected")
	}

	sender.LastSeen = time.Now()

	for name, client := range s.clients {
		if name != args.Username {
			client.Queue = append(client.Queue, Message{
				Sender: args.Username,
				Text:   text,
			})
		}
	}

	*reply = "Message sent"
	return nil
}

func (s *ChatService) Poll(args UserArgs, reply *[]Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[args.Username]
	if !exists {
		return errors.New("user is not connected")
	}

	client.LastSeen = time.Now()
	*reply = append([]Message(nil), client.Queue...)
	client.Queue = nil

	return nil
}

func (s *ChatService) ListUsers(args Empty, reply *[]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make([]string, 0, len(s.clients))

	for username := range s.clients {
		users = append(users, username)
	}

	*reply = users
	return nil
}

func (s *ChatService) Leave(args UserArgs, reply *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[args.Username]; !exists {
		*reply = "User was not connected"
		return nil
	}

	delete(s.clients, args.Username)

	for _, client := range s.clients {
		client.Queue = append(client.Queue, Message{
			Sender: "SERVER",
			Text:   args.Username + " left the chat",
		})
	}

	*reply = "Goodbye, " + args.Username
	return nil
}

func (s *ChatService) CleanupInactive() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for username, client := range s.clients {
		if now.Sub(client.LastSeen) > 15*time.Second {
			delete(s.clients, username)

			for _, other := range s.clients {
				other.Queue = append(other.Queue, Message{
					Sender: "SERVER",
					Text:   username + " disconnected unexpectedly",
				})
			}
		}
	}
}
