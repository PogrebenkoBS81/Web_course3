# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest as builder

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

# Build the Go app
# "-linkmode external -extldflags -static" used for solving problem with sql mattn
# Due to that there is warning during building docker-compose : https://github.com/mattn/go-sqlite3/issues/457
# And as I understand, there is no simple way to handle it.
RUN go build -ldflags "-linkmode external -extldflags -static" -o main .


######## Start a new stage from scratch #######
FROM golang:1.15-alpine  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
# See messsage at the top of docker-compose.yml
COPY config.json /root/config.json

# Expose port 9332 to the outside world
EXPOSE 9332

# Command to run the executable
CMD ["./main"] 