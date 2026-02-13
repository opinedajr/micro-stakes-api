# Micro-Stakes API

Micro-Stakes API is a robust backend solution designed for poker bankroll management and strategy tracking. Built with Go, it provides a secure and scalable platform for users to manage their poker journey, following clean architecture principles and feature-based organization.

## 🚀 Overview

The project aims to provide a comprehensive API for:
- User authentication and profile management (with Keycloak)
- Bankroll tracking and multi-currency support
- Betting strategy analysis
- Detailed session performance metrics

## 🏗 Architecture

This project follows a **Feature-Based Clean Architecture** pattern. Each feature is implemented as an independent module within the `internal/` directory, ensuring high cohesion and low coupling.

### Key Architectural Pillars:
- **Dependency Injection**: All dependencies are managed and injected via a centralized DI container.
- **Interface Segregation**: Strict use of interfaces for service and repository layers.
- **Repository Pattern**: Abstraction of data access logic.
- **Identity Provider**: Integration with Keycloak for secure identity management.
- **Structured Logging**: Using Go's native `slog` for consistent observability.

## 📂 Directory Structure

```text
├── cmd/
│   └── api/                # Application entry point
├── docs/                   # Documentation and samples
├── internal/
│   ├── auth/               # Authentication & User feature module
│   ├── di/                 # Dependency Injection container
│   ├── infrastructure/     # External adapters (Postgres, Keycloak)
│   ├── model/              # Shared domain entities
│   └── shared/             # Shared utilities (config, logger, middleware)
├── migrations/             # Database schema migrations
├── specs/                  # Feature specifications and design documents
├── docker-compose.yml      # Container orchestration
├── Makefile                # Development automation commands
└── .env.sample             # Environment variables template
```

## 🛠 Prerequisites

Ensure you have the following installed:
- **Go**: 1.23+
- **Docker & Docker Compose**: For local infrastructure
- **Make**: For running automation commands
- **golang-migrate**: For manual database migrations (optional)

## ⚙️ Installation & Setup

### 1. Clone the repository
```bash
git clone <repository-url>
cd micro-stakes/backend
```

### 2. Configure Environment Variables
Copy the sample environment file and adjust the values if necessary:
```bash
cp .env.sample .env
```

### 3. Initial Setup (Automated)
The easiest way to get started is using the provided Makefile command:
```bash
make initial-setup
```
This command will:
- Install Go dependencies
- Start PostgreSQL and Keycloak containers
- Run database migrations

## 🐳 Docker Usage

The project uses Docker Compose to manage infrastructure dependencies.

| Service | Port | Description |
|---------|------|-------------|
| **PostgreSQL** | 5432 | Primary database |
| **Keycloak** | 8080 | Identity Provider |
| **API** | 8000 | The Go application (when running in container) |

### Common Docker Commands:
- `make docker-up`: Start all services in background
- `make docker-down`: Stop all services
- `make docker-logs`: View container logs

## 🗄 Database Migrations

Migrations are managed using `golang-migrate`.

- **Run migrations**: `make migrate`
- **Rollback last migration**: `make rollback`
- **Create new migration**: `make migrate-create name=<migration_name>`

## 🧪 Testing

The project maintains a high standard of quality with comprehensive unit and integration tests.

- **Run all tests**: `make test`
- **Run tests with verbose output**: `make test-v`
- **Check test coverage**: `make test-cover`

## 📡 API Endpoints

### Health Check
- **URL**: `GET /health`
- **Description**: Verifies if the API and its dependencies are healthy.

### Authentication

#### User Registration
- **URL**: `POST /auth/register`
- **Description**: Registers a new user in system and Keycloak.
- **Payload**:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "password": "SecurePass123"
}
```
- **Response (201 Created)**:
```json
{
  "id": 1,
  "email": "john.doe@example.com",
  "fullname": "John Doe",
  "message": "User registered successfully"
}
```

#### User Login
- **URL**: `POST /auth/login`
- **Description**: Authenticates user with email and password, returns JWT tokens.
- **Payload**:
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123"
}
```
- **Response (200 OK)**:
```json
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "refreshExpiresIn": 604800
}
```

#### Refresh Token
- **URL**: `POST /auth/refresh`
- **Description**: Exchanges a valid refresh token for new access tokens.
- **Payload**:
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
- **Response (200 OK)**:
```json
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "refreshExpiresIn": 604800
}
```

## 📜 Available Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the API locally |
| `make build` | Build the API binary |
| `make test` | Run tests |
| `make lint` | Run linter (golangci-lint) |
| `make fmt` | Format code |
| `make initial-setup` | Full project bootstrap |
| `make migrate` | Run database migrations |
| `make status` | Check status of project dependencies |

---
