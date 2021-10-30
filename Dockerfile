# syntax=docker/dockerfile:1

FROM arm32v7/golang:1.17.2

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN env GOOS=linux GOARCH=arm GOARM=7 go build -o StarBot

CMD ["./StarBot"]