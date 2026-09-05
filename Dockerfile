# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o telegram-image-workbench main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/telegram-image-workbench .
COPY --from=builder /app/providers.json .

EXPOSE 8080

CMD ["./telegram-image-workbench"]
