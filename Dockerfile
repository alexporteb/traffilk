# Stage 1: Build the React frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:alpine AS go-builder
WORKDIR /app
# Enable CGO for go-sqlite3 and install GCC
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o traffilk-app .

# Stage 3: Final lightweight image
FROM alpine:3.20
WORKDIR /app
# Install ca-certificates and tzdata, sqlite
RUN apk add --no-cache ca-certificates tzdata sqlite

# Copy binary from builder
COPY --from=go-builder /app/traffilk-app .
# Copy compiled frontend assets
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Expose port
EXPOSE 8080

CMD ["./traffilk-app"]
