FROM golang:1.16-alpine

RUN apk add --no-cache git
COPY server.go .
RUN go get -v -d ./...

RUN go build -o httpsink
ENTRYPOINT ./httpsink
