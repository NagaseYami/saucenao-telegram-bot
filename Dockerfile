FROM golang:1.17.7-alpine as builder

WORKDIR /app
COPY . .
RUN go build main.go

FROM alpine

RUN apk update \
    && apk add chromium

COPY --from=builder /app/main /app/main

WORKDIR /app
ENTRYPOINT ["./main"]
