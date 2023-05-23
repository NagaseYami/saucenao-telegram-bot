FROM golang:1.20.4-alpine as builder

WORKDIR /app
COPY . .
RUN go build main.go

FROM alpine

COPY --from=builder /app/main /app/main

WORKDIR /app
ENTRYPOINT ["./main"]
