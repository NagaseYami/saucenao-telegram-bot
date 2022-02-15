FROM golang:1.17.5-alpine as builder

WORKDIR /go/src/app

COPY . .
RUN go build main.go

FROM alpine

COPY --from=builder /go/src/app/main /app/main

WORKDIR /app

CMD ["./main"]
