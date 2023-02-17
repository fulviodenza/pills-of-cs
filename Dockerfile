FROM golang:1.19.5 as builder

# Create and change to the app directory.
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY *.go ./

RUN go build -o /pills

CMD [ "/pills" ]
