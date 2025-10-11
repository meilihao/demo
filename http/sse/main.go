package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Client[T, X comparable] struct {
	Id      T
	Account X
	Channel chan *Data[X]
}

func (c *Client[T, X]) Close() {
	close(c.Channel)
}

type Manager[T, X comparable] struct {
	lock           sync.RWMutex
	clients        map[T]*Client[T, X]
	accounts       map[X]map[T]*Client[T, X]
	Register       chan *Client[T, X]
	Unregister     chan *Client[T, X]
	ForwardMessage chan *Data[X]
}

func NewManager[T, X comparable]() *Manager[T, X] {
	m := &Manager[T, X]{
		clients:        make(map[T]*Client[T, X]),
		accounts:       make(map[X]map[T]*Client[T, X]),
		Register:       make(chan *Client[T, X]),
		Unregister:     make(chan *Client[T, X]),
		ForwardMessage: make(chan *Data[X], 32),
	}

	go m.Run()

	return m
}

func (m *Manager[T, X]) HasClient(client *Client[T, X]) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, ok := m.clients[client.Id]
	return ok
}

func (m *Manager[T, X]) Run() {
	for {
		select {
		case client := <-m.Register:
			m.lock.Lock()

			if _, exists := m.clients[client.Id]; !exists {
				m.clients[client.Id] = client
			}
			if _, exists := m.accounts[client.Account]; !exists {
				m.accounts[client.Account] = make(map[T]*Client[T, X], 3)
			}
			m.accounts[client.Account][client.Id] = client

			slog.Info("client registered", "id", client.Id, "account", client.Account)

			m.lock.Unlock()
		case client := <-m.Unregister:
			m.lock.Lock()

			if c, exists := m.clients[client.Id]; exists {
				delete(m.clients, client.Id)
				delete(m.accounts[client.Account], client.Id)

				c.Close()
			}
			slog.Info("client unregistered", "id", client.Id, "account", client.Account)

			m.lock.Unlock()
		case data := <-m.ForwardMessage:
			m.lock.Lock()

			slog.Info("forward message", "message", data.Message, "from", data.From, "to", data.To)

			if cs, exists := m.accounts[data.To]; exists {
				slog.Info("forwarding message to account", "to", data.To)

				for _, c := range cs {
					c.Channel <- data
				}
			}

			m.lock.Unlock()
		}
	}
}

// Data to be broadcasted to a client.
type Data[X comparable] struct {
	Message string `json:"message"`
	From    X      `json:"sender"`
	To      X      `json:"receiver"`
}

func handleSSE(manager *Manager[string, int]) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set SSE headers
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx

		accountStr := c.Query("account")
		account, _ := strconv.Atoi(accountStr)

		client := &Client[string, int]{
			Id:      c.Query("id"),
			Account: account,
			Channel: make(chan *Data[int], 8),
		}

		// Check for duplicate client IDs
		if manager.HasClient(client) {
			c.JSON(http.StatusConflict, gin.H{"error": "client id already exists"})
			return
		}

		manager.Register <- client

		// Ensure the client is unregistered when the connection closes
		defer func() {
			manager.Unregister <- client
		}()

		// Flush the headers to establish the SSE connection
		fmt.Fprintf(c.Writer, "data: %s\n\n", fmt.Sprintf(`{"event": "time", "data":"%s"}`, time.Now().Format(time.DateTime)))
		c.Writer.Flush()

		// Listen for the connection to close
		closeNotify := c.Writer.CloseNotify()

		// SSE loop: read messages from the client's channel and write them to the response
		for {
			select {
			case data, ok := <-client.Channel:
				slog.Info("received message", "message", data.Message, "from", data.From, "to", data.To)

				// If the channel is closed, exit the loop
				if !ok {
					return
				}

				// Marshal the data to JSON
				payload, err := json.Marshal(data)
				if err != nil {
					fmt.Printf("Error marshaling SSE data: %v\n", err)
					continue
				}

				// Write the SSE formatted message
				// Format: "event: message\ndata: {\"message\":\"...\"}\n\n"
				// Here we use a generic 'message' event.
				if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
					// If writing fails, the client has likely disconnected
					return
				}

				// Flush the data immediately to send it to the client
				c.Writer.Flush()
			case <-closeNotify:
				// Client disconnected, the defer statement will handle unregistration
				return
			}
		}
	}
}

func main() {
	manager := NewManager[string, int]()

	router := gin.Default()
	router.GET("/sse", handleSSE(manager))
	router.StaticFile("/", "./index.html")

	go func() {
		for {
			data := Data[int]{
				From: 0,
				To:   1,
			}

			time.Sleep(5 * time.Second)

			data.Message = fmt.Sprintf("Message from %d to %d @ %v", data.From, data.To, time.Now())

			manager.ForwardMessage <- &data
		}
	}()

	router.Run(":8085")
}
