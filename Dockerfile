FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bot .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates ffmpeg python3 nodejs \
 && wget -qO /usr/local/bin/yt-dlp \
      https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
 && chmod a+rx /usr/local/bin/yt-dlp \
 && echo "--js-runtimes node" > /etc/yt-dlp.conf \
 && adduser -D bigode

WORKDIR /app
COPY --from=builder /app/bot ./bot

USER bigode
ENTRYPOINT ["/app/bot"]