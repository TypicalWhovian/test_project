FROM golang:1.13.5-alpine3.10

RUN apk update
RUN apk --no-cache add gcc g++ make ca-certificates

WORKDIR /go/src/test_project
COPY vendor ../vendor
COPY . ./
