# Phaint Backend
A real-time collaborative drawing application backend built with Go, Firebase, and WebSockets.

## Overview
Phaint is a collaborative digital canvas platform that allows multiple users to create, edit, and share drawing projects in real-time. This backend handles user authentication, project management, real-time collaboration through WebSockets, and persistent storage using Firebase.

## Features

### 🎨 Real-time Collaboration
- Multi-user canvas editing with live cursor tracking
- Real-time synchronization of drawing operations
- User presence indicators with color-coded cursors
- WebSocket-based communication for low-latency updates

### 👥 User Management
- User registration and authentication via Firebase Auth
- Email/password authentication
- User profile management

### 📁 Project Management
- Create and manage drawing projects
- Project collaboration with invite system
- Canvas-based project organization
- Persistent project data storage

### 🎯 Vector Graphics Support
- Multiple drawing tools (paths, rectangles, circles)
- Customizable stroke properties and fills
- Interactive elements with action support
- Canvas background customization

## Architecture

### Core Components

- **WebSocket Handler**: Manages real-time connections and broadcasts
- **Canvas Service**: Thread-safe canvas and vector element management
- **Firebase Integration**: Authentication and data persistence
- **Project System**: Multi-canvas project organization
- **Invitation System**: Secure project sharing

### Data Models

- **Canvas**: Individual drawing surfaces with vector data
- **VectorElements**: Paths, rectangles, and circles with properties
- **Projects**: Collections of canvases with metadata
- **Users**: Authentication and profile information

## Configuration

### Environment Setup

Create `config/secrets/config.yaml`:

```yaml
firebase:
  database_url: "your-firebase-database-url"
  credential_path: "\\config\\secrets\\firebase-credentials.json"
  web_api_key: "your-firebase-web-api-key"
```

### Firebase Setup

1. Create a Firebase project
2. Enable Firestore Database
3. Enable Authentication with Email/Password
4. Download service account credentials
5. Place credentials in `config/secrets/firebase-credentials.json`

### Collections Structure

#### `users`
```json
{
  "UID": "string",
  "mail": "string", 
  "username": "string"
}
```

#### `projects`
```json
{
  "UID": "string",
  "PID": "string",
  "ProjectName": "string",
  "CreationDate": "string",
  "Collaborators": ["string"],
  "CanvasesData": [Canvas]
}
```

#### `invitations`
```json
{
  "CreatorUID": "string",
  "Link": "string",
  "ProjectID": "string",
  "Used": boolean
}
```

## Installation & Setup

### Prerequisites
- Go 1.19+
- Firebase project with Firestore enabled
- Firebase service account credentials

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd phaint_be
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up configuration files (see Configuration section)

4. Run the application:
```bash
go run main.go
```

The server will start on port 8080.

## Development

### Project Structure
```
phaint_be/
├── config/             # Configuration management
├── internal/
│   ├── handlers/      # HTTP and WebSocket handlers
│   ├── services/      # Business logic services
│   └── utils/         # Utility functions
├── models/            # Data models
└── main.go           # Application entry point
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit a pull request

## Support

For issues and questions, please open an issue on the repository or contact the development team.