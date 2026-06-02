FROM golang:1.20-alpine AS builder

WORKDIR /app

# Enable CGO for go-sqlite3 and install GCC
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o traffilk-app .

FROM alpine:latest

WORKDIR /app

# Install ca-certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata sqlite

# Copy binary from builder
COPY --from=builder /app/traffilk-app .
# Copy UI files
COPY ui/ ui/

# Expose port
EXPOSE 8080

CMD ["./traffilk-app"]
