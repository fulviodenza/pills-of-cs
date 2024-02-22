FROM golang:1.19.5 as builder

WORKDIR /app
ADD . /app

RUN go version
RUN go env

RUN go get -u all
RUN go mod tidy

RUN go build -o /pills

CMD [ "/pills" ]
