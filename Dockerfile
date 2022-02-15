FROM golang:1.17.7-bullseye

RUN echo 'deb http://ftp.us.debian.org/debian bullseye main' > /etc/apt/sources.list
RUN apt-get update \
    && apt-get install -y chromium

WORKDIR /app
COPY . .
RUN go build main.go

ENTRYPOINT ["./main"]
