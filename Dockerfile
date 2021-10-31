# syntax=docker/dockerfile:1

FROM arm32v7/golang:1.17.2

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["StarBot"]