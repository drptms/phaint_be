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
	clients        map[*Client]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	users          map[string]*UserPresence
	mutex          sync.RWMutex
	projectID      string
	workBoard      *services.CanvasService
	projectHandler *ProjectHandler
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	username string
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CanvasEvent struct {
	CanvasId string `json:"canvasId"`
	Position Point  `json:"position"`
}

type UserPresence struct {
	UserID    string      `json:"userId"`
	Cursor    *Point      `json:"cursor"`
	Color     string      `json:"color"`
	Username  string      `json:"username"`
	IsDrawing bool        `json:"isDrawing"`
	LastSeen  CanvasEvent `json:"lastSeen"`
}

type Message struct {
	Type      string      `json:"type"`
	Subtype   string      `json:"subtype,omitempty"`
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
	username := r.URL.Query().Get("username")

	if projectID == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	// Get or create hub for this project
	hub := getOrCreateHub(projectID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
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
	docRef, err := GetProjectById(hub.projectID)
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
		broadcast:      make(chan []byte, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		users:          make(map[string]*UserPresence),
		projectID:      projectID,
		workBoard:      services.NewCanvasService(),
		projectHandler: &ProjectHandler{},
	}

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
		Username: client.username,
		LastSeen: CanvasEvent{},
	}
	log.Printf("Client %s connected to project %s. Total clients: %d", client.userID, h.projectID, len(h.clients))
	// Send current users state
	usersData, _ := json.Marshal(Message{Type: "users_state", Data: h.users})

	h.broadcast <- usersData
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.users, client.userID)
		close(client.send)
		log.Printf("Client %s disconnected from project %s", client.userID, h.projectID)
		usersData, _ := json.Marshal(Message{Type: "users_state", Data: h.users})
		h.broadcast <- usersData
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	var msg Message
	if err := json.Unmarshal(message, &msg); err == nil {
		switch msg.Type {
		case "operation":
			h.handleOperations(msg)
		case "users_state":
		case "cursor_move":
			break
		default:
			log.Printf("Unknown message type: %s", msg.Type)
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

func (h *Hub) handleOperations(msg Message) {
	switch msg.Subtype {
	case "load":
		h.handleDrawingOperation(msg)
	case "shape":
		h.handleSingleStroke(msg)
	case "canvas":
		h.handleCanvasBackground(msg)
	case "add":
		h.handleDrawingOperation(msg)
	case "remove":
		h.handleRemoveCanvas(msg)
	case "action":
		h.handleAddAction(msg)
	default:
		log.Printf("Unknown operation subtype: %s", msg.Subtype)
	}
}

func (h *Hub) handleCanvasBackground(msg Message) {
	var canvasId string
	var background string
	if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			canvasId = id
		}
		if backgroundData, ok := dataMap["background"].(string); ok {
			background = backgroundData
		}
	}
	h.workBoard.UpdateCanvasBackground(canvasId, background)
}

func (h *Hub) handleRemoveCanvas(msg Message) {
	var canvasId string
	if dataMap, ok := msg.Data.(string); ok {
		canvasId = dataMap
		h.workBoard.RemoveCanvas(canvasId)
	}
}

func (h *Hub) handleSingleStroke(msg Message) {
	var canvasId string
	var stroke services.VectorElement
	if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			canvasId = id
		}
		if strokeData, ok := dataMap["stroke"].(map[string]interface{}); ok {
			stroke = services.ParseSingleStrokeFromRaw(strokeData)
		}
	}
	h.workBoard.UpdateCanvasElement(canvasId, stroke)

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

	canvas.VectorData.Elements = services.ParseVectorElementsFromRaw(dataMap)

	h.workBoard.AddOrUpdateCanvas(canvas)
}

func (h *Hub) handleAddAction(msg Message) {
	if actionData, ok := msg.Data.(map[string]interface{}); ok {
		var canvasId string
		var vectorElementId string
		var action services.Action
		if id, ok := actionData["canvasId"].(string); ok {
			canvasId = id
		}
		if id, ok := actionData["vectorElementId"].(string); ok {
			vectorElementId = id
		}
		if act, ok := actionData["action"].(map[string]interface{}); ok {
			action = services.Action{
				Type: act["type"].(string),
				Link: act["link"].(string),
			}
		}
		h.workBoard.UpdateCanvasWithAction(canvasId, vectorElementId, action)
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
			c.hub.projectHandler.updateProjectCanvasesData(c.hub)
		}
	}
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}
