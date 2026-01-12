# SecretKeeper

A secure client-server application for managing sensitive data including passwords, credentials, and files. Built with Go and gRPC, featuring end-to-end encryption, JWT authentication, and offline cache support.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Server](#server)
  - [Client](#client)
- [API Reference](#api-reference)
- [Development](#development)
- [Project Structure](#project-structure)
- [Security](#security)
- [License](#license)

## Features

- **Secure Authentication**: JWT-based authentication with access and refresh tokens
- **Secret Management**: Store, retrieve, list, and delete sensitive data
- **File Storage**: Support for file uploads using MinIO object storage
- **Offline Mode**: Client-side caching with fallback when server is unavailable
- **gRPC Communication**: High-performance bidirectional streaming
- **TLS Encryption**: Secure communication between client and server
- **Database Persistence**: PostgreSQL for metadata and user management
- **Containerized Deployment**: Docker and Docker Compose support

## Architecture

SecretKeeper follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                         Client CLI                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Commands   │  │   UseCase    │  │   Adapters   │      │
│  │  (Cobra CLI) │──│   (Business) │──│ (gRPC/Cache) │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │ gRPC/TLS
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Server (gRPC)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Transport   │  │   Service    │  │  Repository  │      │
│  │   (gRPC)     │──│  (Business)  │──│ (Data Layer) │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        ┌──────────┐              ┌──────────┐
        │PostgreSQL│              │  MinIO   │
        │    DB    │              │ Storage  │
        └──────────┘              └──────────┘
```

**Components:**

- **Client**: CLI application built with Cobra for user interaction
- **Server**: gRPC server handling authentication and secret management
- **PostgreSQL**: Stores user credentials and secret metadata
- **MinIO**: Object storage for file-based secrets
- **TLS**: Encrypted communication layer

## Prerequisites

- **Go**: 1.25.2 or higher
- **Docker**: 20.10 or higher
- **Docker Compose**: 2.0 or higher
- **Make**: For running build commands (optional)
- **OpenSSL**: For generating TLS certificates

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/levinOo/SecretKeeper.git
cd SecretKeeper
```

### 2. Generate TLS Certificates

Create a `certs` directory and generate self-signed certificates:

```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

### 3. Configure Environment Variables

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# PostgreSQL Configuration
POSTGRES_USER=secretkeeper
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=secretkeeper_db
POSTGRES_INTERNAL_PORT=5432
POSTGRES_EXTERNAL_PORT=5432

# MinIO Configuration
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_EXTERNAL_PORT=9000
MINIO_INTERNAL_PORT=9000

# Server Configuration
SERVER_CONFIG_PATH=/root/.env
CLIENT_CONFIG_PATH=/root/.env
SERVER_INTERNAL_PORT=50051
SERVER_EXTERNAL_PORT=50051
```

### 4. Configure Server Environment

Create `cmd/server/.env`:

```env
# Database
DB_HOST=db
DB_PORT=5432
DB_USER=secretkeeper
DB_PASSWORD=your_secure_password
DB_NAME=secretkeeper_db
DB_SSL_MODE=disable
DB_CONTEXT_TIMEOUT=30s

# gRPC Server
GRPC_SERVER_ADDR=:50051

# JWT Configuration
JWT_SIGNING_KEY=your_jwt_signing_key_min_32_chars
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Password Security
PEPPER=your_pepper_secret_key

# MinIO Storage
MINIO_ENDPOINT=minio:9000
MINIO_BUCKET=secrets
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MAX_FILE_SIZE=104857600
MAX_SECRET_SIZE=65536
```

### 5. Configure Client Environment

Create `cmd/client/.env`:

```env
CLIENT_SERVER_ADDR=server:50051
CLIENT_CERT_PATH=/root/certs/server.crt
```

### 6. Build and Run with Docker Compose

```bash
docker-compose up --build
```

This will start:
- PostgreSQL database on port 5432
- MinIO object storage on port 9000 (API) and 9001 (Console)
- gRPC server on port 50051
- Client container (interactive mode)

## Configuration

### Server Configuration

The server reads configuration from environment variables defined in `cmd/server/.env`:

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_SERVER_ADDR` | gRPC server listen address | `:50051` |
| `DB_HOST` | PostgreSQL host | - |
| `DB_PORT` | PostgreSQL port | - |
| `DB_USER` | Database user | - |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | - |
| `JWT_SIGNING_KEY` | Secret key for JWT signing | - |
| `JWT_ACCESS_TTL` | Access token lifetime | `15m` |
| `JWT_REFRESH_TTL` | Refresh token lifetime | `168h` |
| `PEPPER` | Additional password hashing secret | - |
| `MINIO_ENDPOINT` | MinIO server endpoint | - |
| `MINIO_BUCKET` | Storage bucket name | - |
| `MAX_FILE_SIZE` | Maximum file size in bytes | `104857600` (100MB) |

### Client Configuration

The client can be configured via environment variables or command-line flags:

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `-s, --server_addr` | `CLIENT_SERVER_ADDR` | gRPC server address |
| `-c, --cert` | `CLIENT_CERT_PATH` | Path to TLS certificate |

## Usage

### Server

The server starts automatically when running `docker-compose up`. It performs:

1. Database migration
2. MinIO bucket initialization
3. gRPC server startup with TLS

**Manual server start:**

```bash
go run cmd/server/main.go
```

### Client

The client provides a CLI interface for interacting with the server.

#### Authentication

**Register a new user:**

```bash
docker-compose exec client ./client register username password
```

**Login:**

```bash
docker-compose exec client ./client login username password
```

#### Secret Management

**Create a secret:**

```bash
# Text secret
docker-compose exec client ./client create-text "my-password" --name "GitHub Token" --description "Personal access token"

# File secret
docker-compose exec client ./client create-file /path/to/file --name "SSH Key" --description "Production server key"
```

**List all secrets:**

```bash
docker-compose exec client ./client list
```

**Get a secret:**

```bash
docker-compose exec client ./client get <secret-id>
```

**Delete a secret:**

```bash
docker-compose exec client ./client delete <secret-id>
```

#### Offline Mode

When the server is unavailable, the client automatically offers to use cached secrets:

```bash
$ docker-compose exec client ./client get abc-123

⚠️  Сервер недоступен. Секрет 'abc-123' найден в локальном кеше.
Использовать кешированную версию секрета? (y/n): y

✅ Секрет успешно получен!
ID: abc-123
Тип: text
...
```

#### Health Check

```bash
docker-compose exec client ./client ping
```

## API Reference

### Authentication Service

#### Register

```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse);
```

**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "access_token": "string",
  "refresh_token": "string",
  "status": "string"
}
```

#### Login

```protobuf
rpc Login(LoginRequest) returns (LoginResponse);
```

**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "access_token": "string",
  "refresh_token": "string",
  "status": "string"
}
```

### Secret Service

#### CreateSecret

```protobuf
rpc CreateSecret(stream CreateSecretRequest) returns (CreateSecretResponse);
```

Streaming upload for large files.

#### GetSecret

```protobuf
rpc GetSecret(GetSecretRequest) returns (stream GetSecretResponse);
```

Streaming download for large files.

#### ListSecrets

```protobuf
rpc ListSecrets(ListSecretsRequest) returns (ListSecretsResponse);
```

#### DeleteSecret

```protobuf
rpc DeleteSecret(DeleteSecretRequest) returns (DeleteSecretResponse);
```

## Development

### Generate Protocol Buffers

```bash
make gen-proto
```

Or manually:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/project.proto
```

### Run Tests

```bash
go test ./...
```

### Database Migrations

Migrations are located in `internal/server/migrations/` and run automatically on server startup using goose.

**Manual migration:**

```bash
goose -dir internal/server/migrations postgres "postgresql://user:password@localhost:5432/dbname" up
```

### Local Development Without Docker

1. Start PostgreSQL and MinIO locally
2. Configure `cmd/server/.env` with local connection strings
3. Run the server:
   ```bash
   CONFIG_PATH=cmd/server/.env go run cmd/server/main.go
   ```
4. Run the client:
   ```bash
   CONFIG_PATH=cmd/client/.env go run cmd/client/main.go <command>
   ```

## Project Structure

```
SecretKeeper/
├── api/                      # Protocol Buffer definitions
│   └── project.proto
├── cmd/                      # Application entry points
│   ├── client/
│   │   └── main.go
│   └── server/
│       └── main.go
├── internal/
│   ├── client/               # Client implementation
│   │   ├── adapter/          # External service adapters (gRPC, cache)
│   │   ├── app/              # Application layer (CLI commands)
│   │   ├── config/           # Configuration management
│   │   ├── domain/           # Domain models
│   │   └── usecase/          # Business logic
│   ├── server/               # Server implementation
│   │   ├── app/              # Application initialization
│   │   ├── config/           # Configuration management
│   │   ├── db/               # Database connection
│   │   ├── domain/           # Domain models
│   │   ├── migrations/       # Database migrations
│   │   ├── repository/       # Data access layer
│   │   ├── service/          # Business logic
│   │   └── transport/        # gRPC handlers
│   └── proto/                # Generated protobuf code
├── pkg/                      # Shared packages
│   ├── client/               # Client utilities
│   └── server/               # Server utilities
├── certs/                    # TLS certificates
├── docker-compose.yml        # Container orchestration
├── Dockerfile                # Multi-stage build
├── Makefile                  # Build automation
└── README.md                 # This file
```

## Security

### Authentication

- Passwords are hashed using bcrypt with a configurable pepper
- JWT tokens are used for stateless authentication
- Refresh tokens allow secure token renewal without re-authentication

### Communication

- All client-server communication is encrypted using TLS
- gRPC provides efficient binary serialization

### Storage

- Sensitive data is stored encrypted in MinIO
- Database credentials are never logged
- Environment variables are used for secrets management

### Best Practices

1. **Change default credentials** in production
2. **Use strong JWT signing keys** (minimum 32 characters)
3. **Rotate secrets regularly**
4. **Enable SSL for MinIO** in production
5. **Use proper certificate authorities** instead of self-signed certificates
6. **Implement rate limiting** for authentication endpoints
7. **Enable database SSL mode** in production

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

**Author**: levinOo  
**Repository**: [https://github.com/levinOo/SecretKeeper](https://github.com/levinOo/SecretKeeper)
