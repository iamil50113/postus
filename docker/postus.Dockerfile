FROM golang:1.22-alpine

WORKDIR /usr/local/src

RUN apk --no-cache add bash gcc gettext

COPY ["app/go.mod", "app/go.sum", "./"]

RUN go mod download