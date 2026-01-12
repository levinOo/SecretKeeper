# Builder
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server
RUN go build -o client ./cmd/client

# Runtime base
FROM alpine:3.20 AS runtime
RUN apk --no-cache add ca-certificates

# Server
FROM runtime AS server
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50051
CMD ["./server"]

# Client
FROM runtime AS client
WORKDIR /app
COPY --from=builder /app/client .
CMD ["./client"]
