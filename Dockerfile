FROM golang:alpine

RUN apk update && apk add git

RUN go get github.com/tsg-ut/smtp2http

ENTRYPOINT ["smtp2http"]

WORKDIR /root/
