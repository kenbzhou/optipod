# NOTE: this is not actually used, right now we are going for an in-tree plugin, this would be if we wanted to use 
# webhook to deploy the plugin (so it would be a standalone pod hence the need for a Dockerfile)
FROM golang:1.21 AS build
WORKDIR /app

COPY . .

RUN go mod tidy && go build -o custom-scheduler plugin.go

FROM alpine:latest
WORKDIR /app

COPY --from=build /app/custom-scheduler /app/custom-scheduler
CMD ["/app/custom-scheduler"]