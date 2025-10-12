# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM golang:1.23-alpine as builder

# Install build dependencies
RUN apk add --no-cache git nodejs npm build-base

# Set up Go environment
ENV GOPROXY="https://proxy.golang.org"
ENV GO111MODULE="on"
ENV CGO_ENABLED=1

# Install Buffalo CLI (latest compatible with Buffalo v1.1.3)
RUN go install github.com/gobuffalo/cli/cmd/buffalo@v0.18.14

# Create and set working directory
RUN mkdir -p /app
WORKDIR /app

# Copy package.json and install Node dependencies first (for caching)
COPY package.json .
RUN npm install

# Copy go mod files and download dependencies (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application
COPY . .

# Build the Buffalo application
RUN buffalo build --static -o /bin/app -v --skip-template-validation

# Final stage - minimal alpine image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache bash ca-certificates tzdata

# Set production environment variables
ENV GO_ENV=production
ENV ADDR=0.0.0.0
ENV PORT=8080

WORKDIR /bin/

# Copy the built binary from builder stage
COPY --from=builder /bin/app .

# Cloud Run expects the container to listen on $PORT
EXPOSE 8080

# Run the binary
CMD exec /bin/app
