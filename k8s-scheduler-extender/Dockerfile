FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler-extender .

# Create minimal final image
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/scheduler-extender .

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8888

ENTRYPOINT ["./scheduler-extender"]