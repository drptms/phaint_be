package handlers

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
	clients    		map[*Client]bool
	broadcast  		chan []byte
	register   		chan *Client
	unregister 		chan *Client
	users      		map[string]*UserPresence
	mutex      		sync.RWMutex
	projectID  		string
	workBoard  		*services.CanvasService
	projectHandler 	*ProjectHandler
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

	userID = generateUserID()

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

	workBoard := hub.getCurrentWorkboard()
	jsonData, err := json.Marshal(workBoard)
	if err != nil {
		log.Println("Error marshaling current workboard:", err)
	} else {
		client.send <- jsonData
	}
}

func initializeHubCanvasData(hub *Hub) error {
	docRef, err := hub.projectHandler.getProjectByName(hub.projectID)
	if err != nil {
		return err
	}

	// Get the document snapshot
	docSnap, err := docRef.Get(context.Background())
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

	switch data := canvasesData.(type) {
    case map[string]interface{}:
        hub.processSingleCanvas(data)
    case []interface{}:
        for _, item := range data {
            if canvasMap, ok := item.(map[string]interface{}); ok {
                hub.processSingleCanvas(canvasMap)
            } else {
                log.Printf("handleDrawingOperation: array item is not map: %T", item)
            }
        }
    default:
        log.Printf("handleDrawingOperation: unexpected data type: %T", data)
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
		workBoard:  services.NewCanvasService(),
		projectHandler: &ProjectHandler{},
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

func (h *Hub) getCurrentWorkboard() map[string]interface{} {
    canvases := h.workBoard.GetAllCanvases()
    transformed := make([]map[string]interface{}, 0, len(canvases))
	log.Print(transformed)
    for _, c := range canvases {
        v := c.VectorData
        transformed = append(transformed, map[string]interface{}{
            "id": c.ID,
            "vectorData": map[string]interface{}{
                "width":          v.Width,
                "height":         v.Height,
                "backgroundFill": v.BackgroundFill,
                "elements":       v.MarshalElements(),
                "timestamp":      v.Timestamp,
                "version":        v.Version,
            },
        })
    }

    return map[string]interface{}{
        "type": "operation",
        "data": transformed,
    }
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
		default:
			log.Printf("Unknown message type: %s", msg.Type)
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
    // Marshal entire dataMap back to JSON bytes
    jsonData, err := json.Marshal(dataMap)
    if err != nil {
        log.Printf("Error marshaling dataMap: %v", err)
        return
    }

    var canvas services.Canvas

    // Unmarshal JSON bytes into Canvas struct
    if err := json.Unmarshal(jsonData, &canvas); err != nil {
        log.Printf("Error unmarshaling to Canvas: %v", err)
        return
    }

    // Since Elements is []VectorElement (interface slice), unmarshal won't fill it properly by default.
    // We need to handle Elements specially:

    canvas.VectorData.Elements = services.ParseVectorElementsFromRaw(dataMap)

    h.workBoard.AddOrUpdateCanvas(canvas)
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
	ticker := time.NewTicker(4 * time.Second)
	defer func() {
		c.hub.projectHandler.updateProjectCanvasesData(c.hub)
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
			c.hub.projectHandler.updateProjectCanvasesData(c.hub)
		}
	}
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
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
