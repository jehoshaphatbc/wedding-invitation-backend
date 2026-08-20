# Wedding Invitation Backend API

Go backend API for wedding invitation management with authentication and user management.

## Tech Stack

- **Language**: Go 1.26
- **Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT (JSON Web Tokens)

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration loader
│   │   └── database.go      # Database connection
│   ├── handler/
│   │   ├── auth_handler.go  # Auth endpoints
│   │   └── user_handler.go  # User CRUD endpoints
│   ├── middleware/
│   │   └── auth.go          # JWT & role middleware
│   ├── model/
│   │   └── user.go          # User model
│   ├── repository/
│   │   └── user_repository.go
│   └── service/
│       ├── auth_service.go  # Auth business logic
│       └── user_service.go  # User business logic
├── pkg/
│   ├── auth/
│   │   └── jwt.go           # JWT utilities
│   └── response/
│       └── response.go      # API response helpers
├── .env                     # Environment variables
├── go.mod
├── go.sum
└── Makefile
```

## Setup

### 1. Prerequisites

- Go 1.26+
- PostgreSQL 12+

### 2. Database Setup

Create the database in PostgreSQL:

```sql
CREATE DATABASE wedding_invitation;
```

### 3. Environment Variables

Copy `.env` and update with your database credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=wedding_invitation
JWT_SECRET=your-secret-key
SERVER_PORT=8080
```

### 4. Run

```bash
make run
```

Or:

```bash
go run cmd/server/main.go
```

## API Endpoints

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login |

### Protected Routes (Requires Bearer Token)

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| GET | `/api/v1/profile` | Get current user profile | All |
| GET | `/api/v1/users` | Get all users | Admin |
| GET | `/api/v1/users/:id` | Get user by ID | Admin |
| PUT | `/api/v1/users/:id` | Update user | Admin |
| DELETE | `/api/v1/users/:id` | Delete user | Admin |
| GET | `/api/v1/stats` | Get user statistics | Admin |

## Request Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Get Profile

```bash
curl http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer <your-token>"
```

### Get All Users (Admin)

```bash
curl http://localhost:8080/api/v1/users?page=1&limit=10 \
  -H "Authorization: Bearer <admin-token>"
```

## User Roles

- **admin**: Full access to all endpoints
- **user**: Basic access (profile only)
- **viewer**: Read-only access

## Development

```bash
# Install dependencies
make deps

# Run in development
make run

# Build binary
make build

# Run tests
make test
```
