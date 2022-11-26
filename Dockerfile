FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY ./bot/go.mod ./bot/go.mod
COPY ./bot/go.sum ./bot/go.sum

RUN go mod download

COPY *.go ./

RUN go build -o /pills_of_cs

CMD [ "/pills_of_cs" ]