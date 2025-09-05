package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"phaint/internal/services"
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
	users      map[string]*UserPresence
	mutex      sync.RWMutex
	projectID  string
	workBoard  services.CanvasService
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
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

func initializeHubCanvasData(hub *Hub) error {
    ctx := context.Background()
    client := services.FirebaseDb().GetClient()

    // Firestore collection and document naming assumed: collection "projects", document projectID
    docSnap, err := client.Collection("projects").Doc(hub.projectID).Get(ctx)
    if err != nil {
        return err
    }

    // Assuming your Firestore doc has a field "CanvasesData" which is a slice or map of canvases
    var rawData map[string]interface{}
    if err := docSnap.DataTo(&rawData); err != nil {
        return err
    }

    canvasesData, ok := rawData["CanvasesData"]
    if !ok {
        return fmt.Errorf("CanvasesData field not found")
    }

    // Parse canvasesData (likely a slice of map[string]interface{} or map[string]interface{})
    // into your Canvas structs and add them to the CanvasService inside hub.workBoard

    // Example assuming canvasesData is a slice of maps (adjust according to your exact Firestore data shape)
    if canvasSlice, ok := canvasesData.([]interface{}); ok {
        for _, c := range canvasSlice {
            canvasMap, ok := c.(map[string]interface{})
            if !ok {
                continue
            }
            // Deserialize each canvasMap into your Canvas struct
            var canvas services.Canvas
            // Use mapstructure or manual unmarshaling or json marshal-unmarshal trick:
            // Convert canvasMap back to JSON bytes and then unmarshal to Canvas struct
            b, err := json.Marshal(canvasMap)
            if err != nil {
                continue
            }
            if err := json.Unmarshal(b, &canvas); err != nil {
                continue
            }
            // Add canvas to the service
            hub.workBoard.AddOrUpdateCanvas(canvas)
        }
    }

    return nil
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
		users:      make(map[string]*UserPresence),
		projectID:  projectID,
		workBoard:  *services.NewCanvasService(),
	}

	// Load canvas data from Firestore and initialize CanvasService
    err := initializeHubCanvasData(hub)
    if err != nil {
        log.Printf("Error loading canvas data for project %s: %v", projectID, err)
        // Optionally continue with empty canvas or handle error accordingly
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
	log.Printf("Broadcasting message to %d clients", len(h.clients))

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

	switch data := msg.Data.(type) {
    case map[string]interface{}:
        h.processSingleCanvas(data)
    case []interface{}:
        for _, item := range data {
            if canvasMap, ok := item.(map[string]interface{}); ok {
                h.processSingleCanvas(canvasMap)
            } else {
                log.Printf("handleDrawingOperation: array item is not map: %T", item)
            }
        }
    default:
        log.Printf("handleDrawingOperation: unexpected data type: %T", data)
    }
}

func (h *Hub) processSingleCanvas(dataMap map[string]interface{}) {
    h.workBoard.AddOrUpdateCanvas(services.Canvas{
        ID:         getString(dataMap, "id"),
        VectorData: services.VectorData{
            Width:          getFloat64(dataMap, "width"),
            Height:         getFloat64(dataMap, "height"),
            Elements:       []services.VectorElement{},
            BackgroundFill: getString(dataMap, "backgroundFill"),
            Timestamp:      getString(dataMap, "timestamp"),
            Version:        getString(dataMap, "version"),
        },
    })
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
		case message, ok := <- c.send:
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
