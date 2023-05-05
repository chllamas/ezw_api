# Use the official Fedora 38 OS as the base image
FROM fedora:38

# Install required packages for Go application
RUN dnf update -y && \
    dnf install -y git golang && \
    dnf clean all

# Set the working directory to the root of the Go application
WORKDIR /app

# Copy the entire Go application to the container
COPY . .

# Build the Go application
RUN go build

# Expose port 8080 for the Go application
EXPOSE 8080

# Set the command to start the Go application when the container starts
CMD ["./ezw_api"]

