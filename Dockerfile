FROM golang:alpine

RUN apk update && apk add git

ENV GO111MODULE on

RUN go get github.com/tsg-ut/smtp2http@master

ENTRYPOINT ["smtp2http"]

WORKDIR /root/
