# WebSocket Integration Guide for Frontend Developers

This document provides guidance on how to integrate with the Unisphere Chat WebSocket API for real-time messaging.

## Overview

The Unisphere Chat system provides both REST APIs for standard CRUD operations and WebSocket connections for real-time messaging. This dual approach allows for:

1. **REST API**: Used for loading message history, creating messages with files, and deleting messages
2. **WebSocket**: Used for real-time messaging and receiving updates about new messages or deleted messages

## WebSocket Connection Endpoint

To establish a WebSocket connection for a specific community chat:

```
WS(S)://{base_url}/api/v1/communities/{communityId}/chat/ws
```

- Use `ws://` for non-secure connections (development)
- Use `wss://` for secure connections (production)
- `{communityId}` should be replaced with the ID of the community you want to connect to

### Authentication

WebSocket connections require authentication. You should include your JWT token as a query parameter:

```
ws://localhost:8080/api/v1/communities/123/chat/ws?token=your_jwt_token
```

## Message Format

### Outgoing Messages (Client to Server)

When sending a message from client to server:

```json
{
  "type": "text",
  "content": "Hello, world!",
  "communityId": 123
}
```

Note that `senderId` and `timestamp` will be filled in by the server to prevent spoofing.

### Incoming Messages (Server to Client)

Messages received from the server will have this format:

```json
{
  "type": "text",
  "communityId": 123,
  "senderId": 456,
  "content": "Hello, world!",
  "timestamp": "2023-06-15T14:30:45.123Z",
  "id": 789
}
```

For file messages:

```json
{
  "type": "file",
  "communityId": 123,
  "senderId": 456,
  "content": "Check out this document",
  "fileUrl": "http://example.com/uploads/files/document.pdf",
  "fileId": 101,
  "timestamp": "2023-06-15T14:35:22.456Z",
  "id": 790
}
```

For message deletion notifications:

```json
{
  "type": "delete",
  "communityId": 123,
  "senderId": 456,
  "id": 789,
  "timestamp": "2023-06-15T14:40:10.789Z"
}
```

## Message Types

The following message types are supported:

1. `text` - Simple text message
2. `file` - Message with a file attachment
3. `delete` - Notification that a message has been deleted

## Connection Lifecycle

### Connection Establishment

1. Create a new WebSocket with the appropriate URL including the community ID and token
2. Listen for the `onopen` event to confirm the connection is established
3. If connection fails, check:
   - That the community exists
   - That your token is valid
   - That the user is a participant in the community

### Handling Messages

1. Set up an `onmessage` handler to process incoming messages
2. Parse the message JSON and update your UI accordingly
3. For file messages, display the appropriate UI for the file type (image preview, file download link, etc.)
4. For delete notifications, remove the message with the matching ID from your UI

### Sending Messages

1. Create a message object with the appropriate format
2. Serialize to JSON using `JSON.stringify()`
3. Send using `websocket.send()`

### Error Handling

1. Set up an `onerror` handler to catch WebSocket errors
2. If the connection is lost, implement a reconnection strategy with exponential backoff

### Connection Closure

1. Listen for the `onclose` event to detect disconnections
2. Close the WebSocket connection when navigating away from the chat using `websocket.close()`

## Example Code

Here's a basic example of WebSocket usage in JavaScript:

```javascript
// Connect to WebSocket
const token = "your_jwt_token";
const communityId = 123;
const socket = new WebSocket(`ws://localhost:8080/api/v1/communities/${communityId}/chat/ws?token=${token}`);

// Connection opened
socket.addEventListener('open', (event) => {
  console.log('Connected to WebSocket');
  
  // Send a test message
  const message = {
    type: 'text',
    content: 'Hello from the client!',
    communityId: communityId
  };
  
  socket.send(JSON.stringify(message));
});

// Listen for messages
socket.addEventListener('message', (event) => {
  try {
    const message = JSON.parse(event.data);
    
    switch (message.type) {
      case 'text':
        console.log(`Received text message: ${message.content}`);
        // Update UI with new message
        break;
        
      case 'file':
        console.log(`Received file message: ${message.content}, file: ${message.fileUrl}`);
        // Display file or link in UI
        break;
        
      case 'delete':
        console.log(`Message deleted: ${message.id}`);
        // Remove message from UI
        break;
        
      default:
        console.log(`Unknown message type: ${message.type}`);
    }
  } catch (e) {
    console.error('Error parsing message:', e);
  }
});

// Handle errors
socket.addEventListener('error', (event) => {
  console.error('WebSocket error:', event);
});

// Handle disconnection
socket.addEventListener('close', (event) => {
  if (event.wasClean) {
    console.log(`Connection closed cleanly, code=${event.code}, reason=${event.reason}`);
  } else {
    console.error('Connection died');
    // Implement reconnection logic here
  }
});

// When leaving the chat page
function leaveChatPage() {
  if (socket) {
    socket.close();
  }
}
```

## REST API Integration

For operations like loading chat history, sending files, or deleting messages, use the REST API:

- **GET** `/api/v1/communities/{id}/chat` - Get chat messages
- **GET** `/api/v1/communities/{id}/chat/{messageId}` - Get a specific chat message
- **POST** `/api/v1/communities/{id}/chat/text` - Send a text message
- **POST** `/api/v1/communities/{id}/chat/file` - Send a file message
- **DELETE** `/api/v1/communities/{id}/chat/{messageId}` - Delete a message

## Testing WebSocket Functionality

A test client is available at `/public/ws-chat-test.html` that simulates two users chatting with each other. You can use this to test your WebSocket implementation.

## Best Practices

1. **Connection Management**:
   - Only maintain one WebSocket connection per community chat
   - Implement reconnection with exponential backoff
   - Close connections when no longer needed

2. **Error Handling**:
   - Always handle WebSocket errors and connection failures
   - Provide appropriate feedback to users

3. **Message Processing**:
   - Validate received messages before processing
   - Handle all message types appropriately
   - Update the UI atomically to avoid race conditions

4. **Performance**:
   - Avoid sending large payloads over WebSockets
   - Use the REST API for file uploads and separate that from real-time messaging

5. **Security**:
   - Never trust client-side data
   - Always validate data on the server-side
   - Keep JWT tokens secure

## Troubleshooting

Common issues and solutions:

1. **Connection refused**: Check if the WebSocket server is running and the URL is correct
2. **Authentication failures**: Verify your JWT token is valid and not expired
3. **Messages not received**: Ensure you're connected to the correct community ID
4. **Connection drops**: Implement a reconnection strategy with backoff

## Need Help?

If you encounter issues integrating with the WebSocket API, please contact the backend development team for assistance.