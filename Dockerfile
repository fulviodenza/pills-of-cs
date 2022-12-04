FROM golang:1.19-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./

# Copy local code to the container image.
COPY . ./

RUN go mod vendor
RUN go install ./...
RUN go build -o server
CMD ["./server"]