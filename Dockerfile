# syntax=docker/dockerfile:1

FROM golang:1.17.2-alpine

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["StarBot"]
