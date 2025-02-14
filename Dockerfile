FROM golang:1.24.0-bullseye AS builder

WORKDIR /app

RUN apt update -y
RUN apt install ca-certificates ffmpeg webp -y

COPY go.mod go.sum ./
RUN go mod download

RUN

COPY . .

RUN GOOS=linux go build -o ostt-tg-bot .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
CMD ["./ostt-tg-bot"]
