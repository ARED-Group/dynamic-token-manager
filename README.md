# Dynamic Token Manager

## Description

Dynamic Token Manager is a secure and efficient service for provisioning tokens and managing edge device configurations. It is tailored for dynamic environments requiring robust token distribution and secure device communication.

---

## Features

- Dynamic token generation, distribution, and management
- Secure communication protocols (e.g., HTTPS, JWT)
- Edge device registration and configuration management
- RESTful API endpoints for token provisioning
- Extensible architecture for custom integrations
- Optional Docker containerization for easy deployment

---

## Project Structure

```
cmd/server/main.go          # Application entry point
internal/token/manager.go   # Token provisioning logic
internal/device/provision.go# Device registration & config
internal/config/config.go   # Configuration management
api/handler.go              # API request handlers
api/routes.go               # Routing setup
pkg/utils.go                # Utility functions
Dockerfile                  # Container build file
.gitignore                  # Git ignore file
go.mod/go.sum               # Go dependencies
```

---

## Requirements

- Go 1.19 or later
- Docker (optional, for containerized deployment)

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/gs01han/dynamic-token-manager.git
cd dynamic-token-manager
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Build & Run

```bash
go run ./cmd/server/main.go
```

### 4. Run with Docker

```bash
docker build -t dynamic-token-manager .
docker run -p 8080:8080 dynamic-token-manager
```

---

## API Endpoints (Example)

| Method | Endpoint            | Description           |
|--------|---------------------|-----------------------|
| POST   | /api/token/provide  | Provision new token   |
| GET    | /api/device/status  | Get device status     |
| POST   | /api/device/register| Register new device   |

---

## Security

- All sensitive endpoints require authentication (JWT recommended)
- HTTPS is enforced for all communications
- Secrets/configuration are managed via environment variables or config files

---

## Testing

```bash
go test ./...
```

---

## Contributing

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/YourFeature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/YourFeature`)
5. Open a Pull Request

---

## License

MIT License.

---

## Contact

Maintained by [Henri Nyakarundi](https://github.com/gs01han).
