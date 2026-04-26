# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install gcc and musl-dev for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s" -o tasks-api .

# Stage 2: Runtime
FROM alpine:3.19

# Install runtime dependencies for SQLite
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/tasks-api .

# Create directory for SQLite database
RUN mkdir -p /app/data

# Expose port
EXPOSE 9090

# Run the application
CMD ["./tasks-api"]
