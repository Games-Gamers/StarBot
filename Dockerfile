# syntax=docker/dockerfile:1

FROM golang

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN env GOOS=linux GOARCH=arm GOARM=5 go build -o StarBot

CMD ["./StarBot"]