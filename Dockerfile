FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY server/ ./server/

RUN CGO_ENABLED=0 GOOS=linux go build -o turkey-clock ./server/

FROM alpine:latest

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app/

COPY --from=builder /app/turkey-clock .
COPY --from=builder /app/server/assets ./assets/

CMD ["./turkey-clock"]
