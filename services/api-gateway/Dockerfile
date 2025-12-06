# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY services ./services
WORKDIR /src/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-gateway ./cmd/api-gateway

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/api-gateway /api-gateway
EXPOSE 8080
ENV GATEWAY_ADDR=:8080 \
    IDENTITY_URL=http://identity:8081 \
    FACILITY_URL=http://facility:8082 \
    BOOKING_URL=http://booking:8083 \
    NOTIFICATION_URL=http://notification:8084
ENTRYPOINT ["/api-gateway"]
