package services

import (
	"sync"
	"time"
)

// Point struct for path points
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// VectorShape base struct
type VectorShape struct {
	ID          string  `json:"id"`
	Stroke      string  `json:"stroke"`
	StrokeWidth float64 `json:"strokeWidth"`
	Fill        string  `json:"fill"`
}

// VectorPath struct
type VectorPath struct {
	VectorShape
	Type   string  `json:"type"`
	Points []Point `json:"points"`
}

// VectorRectangle struct
type VectorRectangle struct {
	VectorShape
	Type   string  `json:"type"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// VectorCircle struct
type VectorCircle struct {
	VectorShape
	Type   string  `json:"type"`
	CX     float64 `json:"cx"`
	CY     float64 `json:"cy"`
	Radius float64 `json:"radius"`
}

// VectorElement interface{} to cover the above types
type VectorElement interface{}

// VectorData struct
type VectorData struct {
	Width          float64         `json:"width"`
	Height         float64         `json:"height"`
	BackgroundFill string          `json:"backgroundFill"`
	Elements       []VectorElement `json:"elements"`
	Timestamp      string          `json:"timestamp"`
	Version        string          `json:"version"`
}

// Canvas struct representing one canvas with ID and vector data
type Canvas struct {
	ID         string     `json:"id"`
	VectorData VectorData `json:"vectorData"`
}

// CanvasService manages multiple canvases safely
type CanvasService struct {
	canvases map[string]*Canvas
	mutex    sync.RWMutex
}

// Constructor
func NewCanvasService() *CanvasService {
	return &CanvasService{
		canvases: make(map[string]*Canvas),
	}
}

// GetCanvas by id; returns nil if not exists
func (c *CanvasService) GetCanvas(id string) *Canvas {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.canvases[id]
}

// AddOrUpdateCanvas inserts or updates a canvas by ID
func (c *CanvasService) AddOrUpdateCanvas(canvas Canvas) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.canvases[canvas.ID] = &canvas
}

// RemoveCanvas deletes a canvas from the map
func (c *CanvasService) RemoveCanvas(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.canvases, id)
}

// UpdateCanvasElements updates elements and metadata of a canvas
func (c *CanvasService) UpdateCanvasElements(id string, elements []VectorElement, timestamp string, version string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	canvas, exists := c.canvases[id]
	if !exists {
		return false
	}
	canvas.VectorData.Elements = elements
	canvas.VectorData.Timestamp = timestamp
	canvas.VectorData.Version = version
	return true
}

// UpdateCanvasBackground changes the background fill color
func (c *CanvasService) UpdateCanvasBackground(id string, backgroundFill string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	canvas, exists := c.canvases[id]
	if !exists {
		return false
	}
	canvas.VectorData.BackgroundFill = backgroundFill
	return true
}

// ListCanvasIDs returns all canvas IDs currently present
func (c *CanvasService) ListCanvasIDs() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ids := make([]string, 0, len(c.canvases))
	for id := range c.canvases {
		ids = append(ids, id)
	}
	return ids
}

// Helper for current timestamp string
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
