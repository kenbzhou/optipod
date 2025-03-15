FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler-extender .

# Create a more useful debugging image
FROM alpine:3.18

WORKDIR /app

# Install debugging utilities
RUN apk add --no-cache \
    curl \
    wget \
    busybox-extras \
    bind-tools \
    emacs-nox \
    vim \
    bash \
    jq \
    tcpdump \
    netcat-openbsd \
    strace \
    procps \
    htop

COPY --from=builder /app/scheduler-extender .

EXPOSE 8888

ENTRYPOINT ["./scheduler-extender"]




## BASE IMAGE BELOW (above is for debug only)

# FROM golang:1.21-alpine AS builder

# WORKDIR /app

# # Copy dependency files
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy source code
# COPY . .

# # Build the binary
# RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler-extender .

# # Create minimal final image
# FROM alpine:3.18

# WORKDIR /app

# COPY --from=builder /app/scheduler-extender .

# # Run as non-root user
# RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# USER appuser

# EXPOSE 8888

# ENTRYPOINT ["./scheduler-extender"]