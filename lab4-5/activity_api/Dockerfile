# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:1.16.6-alpine3.14 as builder

# Add Maintainer Info
LABEL maintainer="Vasily_BS81"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN apk update && apk add gcc g++

# Build the Go app
RUN go build -o main .


######## Start a new stage from scratch #######
FROM alpine:3.14

RUN apk --no-cache add ca-certificates

WORKDIR /root

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
# See messsage at the top of docker-compose.yml
COPY config.json config.json

# Expose port 9332 to the outside world
EXPOSE 9332

# Command to run the executable
CMD ["./main"] 