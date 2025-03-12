FROM golang:1.21.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /scheduler-extender

FROM alpine:3.17
WORKDIR /
COPY --from=builder /scheduler-extender /scheduler-extender
ENTRYPOINT ["/scheduler-extender"]