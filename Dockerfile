FROM golang:1.19.5 as builder

# Create and change to the app directory.
WORKDIR /app
ADD . /app

RUN go build -o /pills

CMD [ "/pills" ]
