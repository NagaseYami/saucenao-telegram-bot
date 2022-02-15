FROM golang:1.17.7-alpine as builder

WORKDIR /go/src/app

COPY . .
RUN go build main.go

FROM debian:bullseye-slim

RUN echo 'deb http://ftp.us.debian.org/debian bullseye main' > /etc/apt/sources.list
RUN apt-get update \
    && apt-get install -y chromium

RUN makedir -p /app
COPY --from=builder /go/src/app/main /app/main

WORKDIR /app

ENTRYPOINT ["./main"]
