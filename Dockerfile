# syntax=docker/dockerfile:1

FROM arm32v7/golang:1.17.2-alpine

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["StarBot"]