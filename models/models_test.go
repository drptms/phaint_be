package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// Test User model
func TestNewUserFromRequest(t *testing.T) {
	userData := User{
		Uid:      "test-uid",
		Username: "testuser",
		Mail:     "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	user, err := NewUserFromRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.Uid != "test-uid" {
		t.Errorf("Expected UID 'test-uid', got '%s'", user.Uid)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Mail != "test@example.com" {
		t.Errorf("Expected mail 'test@example.com', got '%s'", user.Mail)
	}
	if user.Password != "password123" {
		t.Errorf("Expected password 'password123', got '%s'", user.Password)
	}
}

func TestNewUserFromRequestInvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/users", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	_, err := NewUserFromRequest(req)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestNewFirebaseAuthUser(t *testing.T) {
	userData := User{
		Uid:      "test-uid",
		Username: "testuser",
		Mail:     "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	authUser, err := NewFirebaseAuthUser(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if authUser == nil {
		t.Fatal("Expected non-nil auth user")
	}
}

// Test Project model
func TestGetProjectFromRequest(t *testing.T) {
	projectData := map[string]interface{}{
		"Uid":          "user-123",
		"ProjectName":  "Test Project",
		"CreationDate": "2023-01-01",
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	project, err := GetProjectFromRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if project.Uid != "user-123" {
		t.Errorf("Expected UID 'user-123', got '%s'", project.Uid)
	}
	if project.ProjectName != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", project.ProjectName)
	}
	if project.CreationDate != "2023-01-01" {
		t.Errorf("Expected creation date '2023-01-01', got '%s'", project.CreationDate)
	}

	if project.Pid == "" {
		t.Error("Expected non-empty PID")
	}
	if len(project.Pid) != 32 {
		t.Errorf("Expected PID length 32, got %d", len(project.Pid))
	}
}

func TestGetProjectFromRequestInvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/projects", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	_, err := GetProjectFromRequest(req)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetUidFromRequest(t *testing.T) {
	uidData := map[string]string{
		"Uid": "test-uid-123",
	}

	jsonData, _ := json.Marshal(uidData)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	uid, err := GetUidFromRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if uid != "test-uid-123" {
		t.Errorf("Expected UID 'test-uid-123', got '%s'", uid)
	}
}

func TestGetUidFromRequestInvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	_, err := GetUidFromRequest(req)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Test Invitation model
func TestGetInvitationFromRequest(t *testing.T) {
	invitationData := Invitation{
		CreatorUid: "creator-123",
		ProjectID:  "project-456",
	}

	jsonData, _ := json.Marshal(invitationData)
	req, _ := http.NewRequest("POST", "/invitations", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	invitation, err := GetInvitationFromRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if invitation.CreatorUid != "creator-123" {
		t.Errorf("Expected creator UID 'creator-123', got '%s'", invitation.CreatorUid)
	}
	if invitation.ProjectID != "project-456" {
		t.Errorf("Expected project ID 'project-456', got '%s'", invitation.ProjectID)
	}

	if invitation.Link == "" {
		t.Error("Expected non-empty Link")
	}
	if len(invitation.Link) != 32 {
		t.Errorf("Expected Link length 32, got %d", len(invitation.Link))
	}
}

func TestGetInvitationFromRequestInvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/invitations", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	_, err := GetInvitationFromRequest(req)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetInvitationFromRequestEmptyBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/invitations", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	_, err := GetInvitationFromRequest(req)
	if err == nil {
		t.Error("Expected error for empty body, got nil")
	}
}

// Test edge cases
func TestUserModelValidation(t *testing.T) {
	emptyUser := User{}
	if emptyUser.Uid != "" {
		t.Error("Expected empty UID for zero-value User")
	}

	partialUser := User{
		Mail: "test@example.com",
	}
	if partialUser.Username != "" {
		t.Error("Expected empty username for partial User")
	}
}

func TestProjectModelValidation(t *testing.T) {
	emptyProject := Project{}
	if emptyProject.Pid != "" {
		t.Error("Expected empty PID for zero-value Project")
	}

	partialProject := Project{
		ProjectName: "Test Project",
	}
	if partialProject.Uid != "" {
		t.Error("Expected empty UID for partial Project")
	}
}

func TestInvitationModelValidation(t *testing.T) {
	emptyInvitation := Invitation{}
	if emptyInvitation.Link != "" {
		t.Error("Expected empty Link for zero-value Invitation")
	}

	partialInvitation := Invitation{
		CreatorUid: "creator-123",
	}
	if partialInvitation.ProjectID != "" {
		t.Error("Expected empty ProjectID for partial Invitation")
	}
}
