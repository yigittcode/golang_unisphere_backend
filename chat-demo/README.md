# Unisphere Community Chat Demo

A real-time WebSocket chat demonstration showing two users communicating through the Unisphere backend.

## Features

- **Dual User Interface**: Side-by-side chat windows for two simulated users (Alice and Bob)
- **Real-time Messaging**: WebSocket-based real-time communication
- **Authentication**: JWT token-based authentication for both users
- **Modern UI**: Responsive design with gradients and animations
- **Connection Management**: Easy connect/disconnect controls
- **Message Logging**: Real-time activity log for debugging
- **Mobile Responsive**: Works on both desktop and mobile devices

## Setup Instructions

### Prerequisites

1. **Unisphere Backend**: Make sure the Unisphere backend server is running
2. **User Accounts**: You need two registered user accounts with valid JWT tokens
3. **Community**: A community where both users are participants

### Getting JWT Tokens

1. Start your Unisphere backend server
2. Register two users via the `/api/v1/auth/register` endpoint
3. Login both users via the `/api/v1/auth/login` endpoint to get JWT tokens
4. Make sure both users join the same community via `/api/v1/communities/{id}/participants`

### Running the Demo

1. Open `index.html` in a web browser
2. Enter the JWT tokens for both users in the control panel
3. Set the Community ID (default is 1)
4. Verify the WebSocket URL is correct (default: `ws://localhost:8080/api/v1/communities/1/chat/ws`)
5. Click "Connect Both Users"
6. Start chatting! Messages sent by one user will appear in both chat windows

## Configuration

### WebSocket URL Format
```
ws://localhost:8080/api/v1/communities/{communityId}/chat/ws?token={jwt_token}
```

### Default Settings
- **WebSocket URL**: `ws://localhost:8080/api/v1/communities/1/chat/ws`
- **Community ID**: 1
- **User IDs**: Alice (1), Bob (2)

## Message Format

The demo sends and receives messages in the following JSON format:

### Outgoing (Client to Server)
```json
{
  "type": "text",
  "content": "Hello, world!",
  "communityId": 1
}
```

### Incoming (Server to Client)
```json
{
  "type": "text",
  "communityId": 1,
  "senderId": 123,
  "content": "Hello, world!",
  "timestamp": "2023-06-15T14:30:45.123Z",
  "id": 456
}
```

## User Interface

### Alice Johnson (User 1)
- **Avatar**: Red gradient circle with "A1"
- **Role**: Computer Science Student
- **Color Scheme**: Red/Orange gradient

### Bob Smith (User 2)
- **Avatar**: Teal gradient circle with "B2" 
- **Role**: Engineering Student
- **Color Scheme**: Teal/Green gradient

## Connection Status

- 🔴 **Disconnected**: User is not connected to WebSocket
- 🟡 **Connecting**: WebSocket connection in progress
- 🟢 **Connected**: User is connected and can send/receive messages

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Make sure the backend server is running
   - Verify the WebSocket URL is correct
   - Check if the port (8080) is accessible

2. **Authentication Failed**
   - Verify JWT tokens are valid and not expired
   - Make sure users are registered in the system
   - Check if users are participants in the specified community

3. **Messages Not Appearing**
   - Confirm both users are connected to the same community
   - Check the browser console for JavaScript errors
   - Verify the Community ID is correct

4. **WebSocket Errors**
   - Check the log panel at the bottom for detailed error messages
   - Ensure CORS is properly configured on the backend
   - Verify the WebSocket endpoint is accessible

## Technical Details

### Technologies Used
- **HTML5**: Structure and WebSocket API
- **CSS3**: Modern styling with flexbox and gradients
- **Vanilla JavaScript**: WebSocket handling and DOM manipulation
- **WebSocket**: Real-time bidirectional communication

### Browser Compatibility
- Chrome 16+
- Firefox 11+
- Safari 7+
- Edge 12+

### Performance
- Lightweight: No external dependencies
- Efficient: Event-driven message handling
- Scalable: Can handle multiple concurrent connections

## Development

### File Structure
```
chat-demo/
├── index.html          # Main demo page
└── README.md          # This documentation
```

### Customization

To customize the demo:

1. **Change User Information**: Modify the user names, avatars, and roles in the HTML
2. **Update Colors**: Adjust the CSS gradient colors for different themes  
3. **Add Features**: Extend the JavaScript to support file messages, message deletion, etc.
4. **Styling**: Modify the CSS for different layouts or responsive behavior

## License

This demo is part of the Unisphere project and follows the same licensing terms.