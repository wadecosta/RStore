# === STAGE 1: Build Environment ===
FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code and compile statically
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# === STAGE 2: Lightweight Runtime ===
FROM docker.io/library/alpine:latest

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./server"]
