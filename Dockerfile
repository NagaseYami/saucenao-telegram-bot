FROM golang:1.17.7-alpine as builder

WORKDIR /go/src/app

COPY . .
RUN go build main.go

FROM debian:bullseye-slim

RUN apt-get update \
    && apt-get install -y chromium-browser

COPY --from=builder /go/src/app/main /app/main

WORKDIR /app

CMD ["./main"]
