# UniSphere Backend

A Go-based backend API for the UniSphere application.

## Features

- User authentication with JWT
- Email verification system
- RESTful API design
- Faculty, department and course management
- Upload system for class notes and past exams
- Community features
- Role-based access control
- Swagger documentation

## Prerequisites

- Go 1.20+
- PostgreSQL 14+
- Docker and Docker Compose (optional)

## Getting Started

### Local Development

1. Clone the repository:

```bash
git clone https://your-repository-url.git
cd golang_unisphere_backend
```

2. Install dependencies:

```bash
go mod download
```

3. Configure the application:

Copy the example configuration file:

```bash
cp configs/configExample.yaml configs/config.yaml
```

Update the configuration file with your settings.

4. Run migrations:

```bash
go run cmd/migrations/main.go
```

5. Start the application:

```bash
make dev
```

### Using Docker

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Update the `.env` file with your settings.

3. Build and start the containers:

```bash
docker-compose up -d
```

The API will be available at http://localhost:8080.

## API Documentation

Swagger documentation is available at `/swagger/index.html` when the application is running.

## Email Verification

The system includes email verification for user registration:

1. When a user registers, they receive a verification email with a token
2. The user must verify their email before they can use most features of the application
3. User profile endpoints are accessible without email verification, allowing users to update their profile while verification is pending

For testing without SMTP configuration, verification tokens are logged to the console.

## Project Structure

- `cmd/api`: Application entry point
- `internal/app`: Core application code
  - `controllers`: HTTP request handlers
  - `models`: Data models and DTOs
  - `repositories`: Database access layer
  - `services`: Business logic
  - `routes`: URL routing
- `internal/pkg`: Shared packages and utilities
  - `auth`: Authentication utilities
  - `email`: Email service
  - `validation`: Input validation
- `migrations`: SQL migrations
- `configs`: Configuration files
- `docs`: Swagger documentation

## License

[Your License]