package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct{}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	operations []DrawingOperation
	users      map[string]*UserPresence
	mutex      sync.RWMutex
	projectID  string
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type DrawingOperation struct {
	Type      string  `json:"type"`
	Tool      string  `json:"tool"`
	Color     string  `json:"color"`
	Points    []Point `json:"points"`
	Timestamp int64   `json:"timestamp"`
	UserID    string  `json:"userId"`
	ID        string  `json:"id"`
	ProjectID string  `json:"projectId"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type UserPresence struct {
	UserID    string    `json:"userId"`
	Cursor    *Point    `json:"cursor"`
	Color     string    `json:"color"`
	IsDrawing bool      `json:"isDrawing"`
	LastSeen  time.Time `json:"lastSeen"`
}

type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"userId,omitempty"`
	ProjectID string      `json:"projectId,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development only
	},
}

// Global hubs for different projects
var projectHubs = make(map[string]*Hub)
var hubsMutex = sync.RWMutex{}

func (wh *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectId")
	userID := r.URL.Query().Get("userId")

	if projectID == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	if userID == "" {
		userID = generateUserID()
	}

	// Get or create hub for this project
	hub := getOrCreateHub(projectID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func getOrCreateHub(projectID string) *Hub {
	hubsMutex.Lock()
	defer hubsMutex.Unlock()

	if hub, exists := projectHubs[projectID]; exists {
		return hub
	}

	hub := &Hub{
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		operations: make([]DrawingOperation, 0),
		users:      make(map[string]*UserPresence),
		projectID:  projectID,
	}

	projectHubs[projectID] = hub
	go hub.run()

	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	h.users[client.userID] = &UserPresence{
		UserID:   client.userID,
		Color:    fmt.Sprintf("#%06x", time.Now().UnixNano()%0xFFFFFF),
		LastSeen: time.Now(),
	}

	log.Printf("Client %s connected to project %s. Total clients: %d", client.userID, h.projectID, len(h.clients))

	// Send current drawing operations
	for _, op := range h.operations {
		data, _ := json.Marshal(Message{Type: "operation", Data: op})
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
			return
		}
	}

	// Send current users state
	usersData, _ := json.Marshal(Message{Type: "users_state", Data: h.users})
	select {
	case client.send <- usersData:
	default:
		close(client.send)
		delete(h.clients, client)
	}

}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.users, client.userID)
		close(client.send)
		log.Printf("Client %s disconnected from project %s", client.userID, h.projectID)
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var msg Message
	if err := json.Unmarshal(message, &msg); err == nil {
		switch msg.Type {
		case "operation":
			h.handleDrawingOperation(msg)
		case "cursor_move":
			h.handleCursorMove(msg)
		}
	}

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
			delete(h.users, client.userID)
		}
	}
}

func (h *Hub) handleDrawingOperation(msg Message) {
	if opData, ok := msg.Data.(map[string]interface{}); ok {
		op := parseDrawingOperation(opData)
		op.ProjectID = h.projectID
		h.operations = append(h.operations, op)
	}
}

func (h *Hub) handleCursorMove(msg Message) {
	if cursorData, ok := msg.Data.(map[string]interface{}); ok {
		if user, exists := h.users[msg.UserID]; exists {
			if position, ok := cursorData["position"].(map[string]interface{}); ok {
				user.Cursor = &Point{
					X: getFloat64(position, "x"),
					Y: getFloat64(position, "y"),
				}
				user.LastSeen = time.Now()
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			msg.UserID = c.userID
			msg.ProjectID = c.hub.projectID
			if newMessage, err := json.Marshal(msg); err == nil {
				message = newMessage
			}
		}

		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			_ = c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func parseDrawingOperation(data map[string]interface{}) DrawingOperation {
	points := make([]Point, 0)
	if pointsData, ok := data["points"].([]interface{}); ok {
		for _, p := range pointsData {
			if pointMap, ok := p.(map[string]interface{}); ok {
				points = append(points, Point{
					X: getFloat64(pointMap, "x"),
					Y: getFloat64(pointMap, "y"),
				})
			}
		}
	}

	return DrawingOperation{
		Type:      getString(data, "type"),
		Tool:      getString(data, "tool"),
		Color:     getString(data, "color"),
		Points:    points,
		Timestamp: getInt64(data, "timestamp"),
		UserID:    getString(data, "userId"),
		ID:        getString(data, "id"),
	}
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	return 0
}
