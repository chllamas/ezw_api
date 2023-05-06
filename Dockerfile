# Use the official Fedora 38 OS as the base image
FROM alpine:latest

# Install required packages for Go application
RUN apk add --no-cache git go

# Set the working directory to the root of the Go application
WORKDIR /ezw_api

# Copy the entire Go application to the container
COPY . .

# Build the Go application
RUN go build -o ezw_api .

# Expose port 8080 for the Go application
EXPOSE 8080

# Set the command to start the Go application when the container starts
CMD ["./ezw_api"]

