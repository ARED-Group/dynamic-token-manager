FROM golang:1.21-alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# Create final minimal image
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/server /app/

# Copy any additional configuration files
COPY .env.example /app/.env

EXPOSE 8080

CMD ["/app/server"]