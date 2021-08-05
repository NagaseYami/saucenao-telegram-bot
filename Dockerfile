FROM golang:1.16-alpine

WORKDIR /go/src/app

RUN apt-get update \
    && apt-get install -y git

RUN git clone git@github.com:NagaseYami/saucenao-telegram-bot.git /go/src/app
RUN go build /go/src/app/main.go

CMD ["app"]
