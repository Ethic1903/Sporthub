# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY services ./services
WORKDIR /src/services/notification
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/notification ./cmd/notification

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/notification /notification
EXPOSE 8084
ENV NOTIFICATION_ADDR=:8084
ENTRYPOINT ["/notification"]
