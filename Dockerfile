FROM golang:1.17.7-alpine as builder

WORKDIR /go/src/app

COPY . .
RUN go build main.go

FROM alpine

RUN apk add chromium
COPY --from=builder /go/src/app/main /app/main

WORKDIR /app

CMD ["./main"]
