# Start from the latest golang base image
FROM golang:1.24-alpine3.21 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
WORKDIR /app/cmd/nats-helper
RUN go build -o /nats-helper

FROM alpine:3.20

WORKDIR /app/

COPY --from=builder /nats-helper .

ENTRYPOINT ["./nats-helper"]
CMD [ "--config=./config.yaml" ]
