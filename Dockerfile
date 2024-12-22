FROM golang:1.21 AS builder
LABEL authors="agcon"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/cache-service

FROM ubuntu:22.04

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app /app/app

EXPOSE 8084

ENTRYPOINT ["/app/app"]
