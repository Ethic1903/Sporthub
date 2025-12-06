# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY services ./services
WORKDIR /src/services/booking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/booking ./cmd/booking

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/booking /booking
EXPOSE 8083
ENV BOOKING_ADDR=:8083
ENV FACILITY_SERVICE_URL=http://facility:8082
ENTRYPOINT ["/booking"]
