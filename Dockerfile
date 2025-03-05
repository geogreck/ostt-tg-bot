FROM golang:1.24.0-bullseye AS builder

WORKDIR /app

RUN apt update -y
RUN apt install ca-certificates ffmpeg webp -y

# install and launch google chrome for chromedp usage
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
&& echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list
RUN apt-get update && apt-get -y install google-chrome-stable
RUN chrome &

# install fonts with emoji support
RUN apt-get install -y fonts-noto-color-emoji fonts-noto fonts-noto-cjk fonts-liberation fonts-dejavu \
    && fc-cache -f -v

COPY go.mod go.sum ./
RUN go mod download

RUN

COPY . .

RUN GOOS=linux go build -o ostt-tg-bot .

CMD ["./ostt-tg-bot"]
