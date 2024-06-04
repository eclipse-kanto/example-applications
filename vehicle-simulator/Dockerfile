# Stage 1: Build the Go application
FROM golang:alpine AS builder

# Install git
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /build

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .

# Stage 2: Create the final lightweight image
FROM scratch

# Set the working directory
WORKDIR /source

# Copy application executable from builder
COPY --from=builder /build/app /bin/app

# Command to run the application
CMD ["/bin/app"]
