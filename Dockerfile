FROM golang:1.22.2-alpine AS builder

WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bot .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/bot ./bot

EXPOSE 0

ENTRYPOINT ["/app/bot"]
