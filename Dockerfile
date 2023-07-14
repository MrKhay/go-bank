# Use the official Golang base image
FROM --platform=$BUILDPLATFORM golang:alpine


# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the working directory
COPY go.mod go.sum ./

# Download and cache Go dependencies
RUN go mod download

# Copy the entire project directory to the working directory
COPY . .

# Build the Go application
RUN go build -o /gobank

# Expose the desired port
EXPOSE 3001

# Set the command to run the Go application
CMD ["/gobank"]
