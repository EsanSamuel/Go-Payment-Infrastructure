package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"example.com/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	UserId string
	Conn   *websocket.Conn
	Hub    *Hub
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) HandleWS(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID := fmt.Sprint(user_id)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		UserId: userID,
		Conn:   conn,
		Hub:    h,
	}

	h.addClient(client)

	log.Printf("user %s connected", userID)

	defer h.removeClient(client)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("user %s disconnected: %v", userID, err)
			return
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.clients[client.UserId]; ok {
		existing.Conn.Close()
	}

	h.clients[client.UserId] = client
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	current, ok := h.clients[client.UserId]

	if ok && current == client {
		delete(h.clients, client.UserId)
	}

	client.Conn.Close()

	log.Printf("user %s removed", client.UserId)
}

func (h *Hub) SendNotification(userID string, notification models.NotificationResponse) {
	log.Printf("SendNotification called for userID=%q", userID)
	h.mu.RLock()

	client, ok := h.clients[userID]

	defer h.mu.RUnlock()
	if !ok {
		log.Printf(
			"user %s is offline",
			userID,
		)
		return
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf(
			"failed to marshal notification: %v",
			err,
		)
		return
	}
	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf(
			"failed to send notification to %s: %v",
			userID,
			err,
		)

		h.removeClient(client)
	}
}
