package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Point struct for path points
type Point struct {
	X float64 `firestore:"x" json:"x"`
	Y float64 `firestore:"y" json:"y"`
}

type Action struct {
	Type string `firestore:"type" json:"type"`
	Link string `firestore:"link" json:"link"`
}

// VectorShape base struct
type VectorShape struct {
	ID          string  `firestore:"id" json:"id"`
	Stroke      string  `firestore:"stroke" json:"stroke"`
	StrokeWidth float64 `firestore:"strokeWidth" json:"strokeWidth"`
	Fill        string  `firestore:"fill" json:"fill"`
	Action      Action  `firestore:"action" json:"action"`
}

// VectorPath struct
type VectorPath struct {
	VectorShape
	Type   string  `firestore:"type" json:"type"`
	Points []Point `firestore:"points" json:"points"`
}

// VectorRectangle struct
type VectorRectangle struct {
	VectorShape
	Type   string  `firestore:"type" json:"type"`
	X      float64 `firestore:"x" json:"x"`
	Y      float64 `firestore:"y" json:"y"`
	Width  float64 `firestore:"width" json:"width"`
	Height float64 `firestore:"height" json:"height"`
}

// VectorCircle struct
type VectorCircle struct {
	VectorShape
	Type   string  `firestore:"type" json:"type"`
	CX     float64 `firestore:"cx" json:"cx"`
	CY     float64 `firestore:"cy" json:"cy"`
	Radius float64 `firestore:"radius" json:"radius"`
}

// VectorElement interface{} to cover the above types
type VectorElement interface{}

// VectorData struct
type VectorData struct {
	Width          float64         `firestore:"width" json:"width"`
	Height         float64         `firestore:"height" json:"height"`
	BackgroundFill string          `firestore:"backgroundFill" json:"backgroundFill"`
	Elements       []VectorElement `firestore:"elements" json:"elements"`
	Timestamp      string          `firestore:"timestamp" json:"timestamp"`
	Version        string          `firestore:"version" json:"version"`
}

// Canvas struct representing one canvas with ID and vector data
type Canvas struct {
	ID         string     `firestore:"id" json:"id"`
	VectorData VectorData `firestore:"vectorData" json:"vectorData"`
}

// CanvasService manages multiple canvases safely
type CanvasService struct {
	canvases map[string]*Canvas
	mutex    sync.RWMutex
}

func (c *CanvasService) GetAllCanvases() []*Canvas {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	canvases := make([]*Canvas, 0, len(c.canvases))
	for _, canvas := range c.canvases {
		canvases = append(canvases, canvas)
	}
	return canvases
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

// Special parser for VectorElements with dynamic type handling
func ParseVectorElementsFromRaw(dataMap map[string]interface{}) []VectorElement {
	//log.Print(dataMap)
	vectorData, ok := dataMap["vectorData"].(map[string]interface{})
	if !ok {
		return nil
	}

	if vectorData == nil {
		return nil
	}

	elementsRaw, ok := vectorData["elements"]
	if !ok {
		return nil
	}
	//log.Print(elementsRaw)

	elements := []VectorElement{}
	rawSlice, ok := elementsRaw.([]interface{})
	if !ok {
		return nil
	}

	//log.Print(rawSlice)
	for _, elem := range rawSlice {
		elemMap, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}

		t := getString(elemMap, "type")
		jsonData, err := json.Marshal(elemMap)
		if err != nil {
			log.Println(err)
			continue
		}

		switch t {
		case "path":
			var path VectorPath
			if err := json.Unmarshal(jsonData, &path); err == nil {
				elements = append(elements, path)
			}
		case "rectangle":
			var rect VectorRectangle
			if err := json.Unmarshal(jsonData, &rect); err == nil {
				elements = append(elements, rect)
			}
		case "circle":
			var circle VectorCircle
			if err := json.Unmarshal(jsonData, &circle); err == nil {
				elements = append(elements, circle)
			}
		default:
			log.Printf("Unknown vector element type: %s", t)
		}
	}
	return elements
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (v *VectorData) MarshalElements() []interface{} {
	results := make([]interface{}, len(v.Elements))
	for i, el := range v.Elements {
		results[i] = el // assert concrete type or convert
	}
	return results
}
