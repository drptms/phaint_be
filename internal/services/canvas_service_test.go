package services

import (
	"testing"
)

func TestNewCanvasService(t *testing.T) {
	cs := NewCanvasService()
	if cs == nil {
		t.Fatal("NewCanvasService returned nil")
	}
	if cs.canvases == nil {
		t.Fatal("canvases map is nil")
	}
	if len(cs.canvases) != 0 {
		t.Errorf("Expected empty canvases map, got %d items", len(cs.canvases))
	}
}

func TestAddOrUpdateCanvas(t *testing.T) {
	cs := NewCanvasService()
	canvas := Canvas{
		ID: "test-canvas-1",
		VectorData: VectorData{
			Width:          800,
			Height:         600,
			BackgroundFill: "#ffffff",
			Elements:       []VectorElement{},
			Timestamp:      GetCurrentTimestamp(),
			Version:        "1.0",
		},
	}

	cs.AddOrUpdateCanvas(canvas)
	retrieved := cs.GetCanvas("test-canvas-1")
	if retrieved == nil {
		t.Fatal("Canvas not found after adding")
	}
	if retrieved.ID != "test-canvas-1" {
		t.Errorf("Expected ID 'test-canvas-1', got '%s'", retrieved.ID)
	}
	if retrieved.VectorData.Width != 800 {
		t.Errorf("Expected width 800, got %f", retrieved.VectorData.Width)
	}

	canvas.VectorData.Width = 1000
	cs.AddOrUpdateCanvas(canvas)
	retrieved = cs.GetCanvas("test-canvas-1")
	if retrieved.VectorData.Width != 1000 {
		t.Errorf("Expected updated width 1000, got %f", retrieved.VectorData.Width)
	}
}

func TestGetCanvas(t *testing.T) {
	cs := NewCanvasService()

	canvas := cs.GetCanvas("non-existent")
	if canvas != nil {
		t.Error("Expected nil for non-existent canvas")
	}

	testCanvas := Canvas{
		ID: "test-canvas-2",
		VectorData: VectorData{
			Width:  400,
			Height: 300,
		},
	}
	cs.AddOrUpdateCanvas(testCanvas)

	retrieved := cs.GetCanvas("test-canvas-2")
	if retrieved == nil {
		t.Fatal("Canvas not found")
	}
	if retrieved.ID != "test-canvas-2" {
		t.Errorf("Expected ID 'test-canvas-2', got '%s'", retrieved.ID)
	}
}

func TestRemoveCanvas(t *testing.T) {
	cs := NewCanvasService()

	canvas := Canvas{
		ID:         "test-canvas-3",
		VectorData: VectorData{Width: 100, Height: 100},
	}
	cs.AddOrUpdateCanvas(canvas)

	cs.RemoveCanvas("test-canvas-3")

	if cs.GetCanvas("test-canvas-3") != nil {
		t.Error("Canvas should not exist after removal")
	}
}

func TestUpdateCanvasElement(t *testing.T) {
	cs := NewCanvasService()

	canvas := Canvas{
		ID: "test-canvas-4",
		VectorData: VectorData{
			Elements: []VectorElement{},
		},
	}
	cs.AddOrUpdateCanvas(canvas)

	element := VectorPath{
		VectorShape: VectorShape{
			ID:     "element-1",
			Stroke: "#000000",
		},
		Type:   "path",
		Points: []Point{{X: 10, Y: 10}, {X: 20, Y: 20}},
	}

	success := cs.UpdateCanvasElement("test-canvas-4", element)
	if !success {
		t.Error("UpdateCanvasElement should return true for existing canvas")
	}

	retrieved := cs.GetCanvas("test-canvas-4")
	if len(retrieved.VectorData.Elements) != 1 {
		t.Errorf("Expected 1 element, got %d", len(retrieved.VectorData.Elements))
	}

	success = cs.UpdateCanvasElement("non-existent", element)
	if success {
		t.Error("UpdateCanvasElement should return false for non-existent canvas")
	}
}

func TestUpdateCanvasBackground(t *testing.T) {
	cs := NewCanvasService()

	canvas := Canvas{
		ID: "test-canvas-5",
		VectorData: VectorData{
			BackgroundFill: "#ffffff",
		},
	}
	cs.AddOrUpdateCanvas(canvas)

	success := cs.UpdateCanvasBackground("test-canvas-5", "#ff0000")
	if !success {
		t.Error("UpdateCanvasBackground should return true for existing canvas")
	}

	retrieved := cs.GetCanvas("test-canvas-5")
	if retrieved.VectorData.BackgroundFill != "#ff0000" {
		t.Errorf("Expected background '#ff0000', got '%s'", retrieved.VectorData.BackgroundFill)
	}

	success = cs.UpdateCanvasBackground("non-existent", "#00ff00")
	if success {
		t.Error("UpdateCanvasBackground should return false for non-existent canvas")
	}
}

func TestGetAllCanvases(t *testing.T) {
	cs := NewCanvasService()

	canvases := cs.GetAllCanvases()
	if len(canvases) != 0 {
		t.Errorf("Expected 0 canvases, got %d", len(canvases))
	}

	canvas1 := Canvas{ID: "canvas-1", VectorData: VectorData{Width: 100}}
	canvas2 := Canvas{ID: "canvas-2", VectorData: VectorData{Width: 200}}
	cs.AddOrUpdateCanvas(canvas1)
	cs.AddOrUpdateCanvas(canvas2)

	canvases = cs.GetAllCanvases()
	if len(canvases) != 2 {
		t.Errorf("Expected 2 canvases, got %d", len(canvases))
	}
}

func TestUpdateCanvasWithAction(t *testing.T) {
	cs := NewCanvasService()

	rectangle := VectorRectangle{
		VectorShape: VectorShape{
			ID: "rect-1",
		},
		Type: "rectangle",
		X:    10, Y: 10, Width: 50, Height: 50,
	}

	canvas := Canvas{
		ID: "test-canvas-6",
		VectorData: VectorData{
			Elements: []VectorElement{rectangle},
		},
	}
	cs.AddOrUpdateCanvas(canvas)

	action := Action{
		Type: "link",
		Link: "https://example.com",
	}

	success := cs.UpdateCanvasWithAction("test-canvas-6", "rect-1", action)
	if !success {
		t.Error("UpdateCanvasWithAction should return true for existing canvas and element")
	}

	retrieved := cs.GetCanvas("test-canvas-6")
	if len(retrieved.VectorData.Elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(retrieved.VectorData.Elements))
	}

	if rect, ok := retrieved.VectorData.Elements[0].(VectorRectangle); ok {
		if rect.Action.Type != "link" {
			t.Errorf("Expected action type 'link', got '%s'", rect.Action.Type)
		}
		if rect.Action.Link != "https://example.com" {
			t.Errorf("Expected action link 'https://example.com', got '%s'", rect.Action.Link)
		}
	} else {
		t.Error("Element is not a VectorRectangle")
	}
}

func TestParseVectorElementsFromRaw(t *testing.T) {
	rawData := map[string]interface{}{
		"vectorData": map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{
					"type":   "path",
					"id":     "path-1",
					"stroke": "#000000",
					"points": []interface{}{
						map[string]interface{}{"x": 10.0, "y": 10.0},
						map[string]interface{}{"x": 20.0, "y": 20.0},
					},
				},
				map[string]interface{}{
					"type":   "rectangle",
					"id":     "rect-1",
					"stroke": "#ff0000",
					"x":      50.0,
					"y":      50.0,
					"width":  100.0,
					"height": 80.0,
				},
			},
		},
	}

	elements := ParseVectorElementsFromRaw(rawData)
	if len(elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(elements))
	}

	if path, ok := elements[0].(VectorPath); ok {
		if path.Type != "path" {
			t.Errorf("Expected type 'path', got '%s'", path.Type)
		}
		if path.ID != "path-1" {
			t.Errorf("Expected ID 'path-1', got '%s'", path.ID)
		}
		if len(path.Points) != 2 {
			t.Errorf("Expected 2 points, got %d", len(path.Points))
		}
	} else {
		t.Error("First element is not a VectorPath")
	}

	if rect, ok := elements[1].(VectorRectangle); ok {
		if rect.Type != "rectangle" {
			t.Errorf("Expected type 'rectangle', got '%s'", rect.Type)
		}
		if rect.Width != 100.0 {
			t.Errorf("Expected width 100.0, got %f", rect.Width)
		}
	} else {
		t.Error("Second element is not a VectorRectangle")
	}
}

func TestParseSingleStrokeFromRaw(t *testing.T) {
	rawCircle := map[string]interface{}{
		"type":   "circle",
		"id":     "circle-1",
		"stroke": "#0000ff",
		"cx":     100.0,
		"cy":     100.0,
		"radius": 50.0,
	}

	element := ParseSingleStrokeFromRaw(rawCircle)
	if element == nil {
		t.Fatal("ParseSingleStrokeFromRaw returned nil")
	}

	if circle, ok := element.(VectorCircle); ok {
		if circle.Type != "circle" {
			t.Errorf("Expected type 'circle', got '%s'", circle.Type)
		}
		if circle.Radius != 50.0 {
			t.Errorf("Expected radius 50.0, got %f", circle.Radius)
		}
	} else {
		t.Error("Element is not a VectorCircle")
	}
}

func TestVectorDataMarshalElements(t *testing.T) {
	vd := VectorData{
		Elements: []VectorElement{
			VectorPath{
				VectorShape: VectorShape{ID: "path-1"},
				Type:        "path",
			},
			VectorRectangle{
				VectorShape: VectorShape{ID: "rect-1"},
				Type:        "rectangle",
			},
		},
	}

	marshaled := vd.MarshalElements()
	if len(marshaled) != 2 {
		t.Errorf("Expected 2 marshaled elements, got %d", len(marshaled))
	}

	if path, ok := marshaled[0].(VectorPath); ok {
		if path.ID != "path-1" {
			t.Errorf("Expected ID 'path-1', got '%s'", path.ID)
		}
	} else {
		t.Error("First marshaled element is not a VectorPath")
	}
}
