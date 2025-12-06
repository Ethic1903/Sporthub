# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY services ./services
WORKDIR /src/services/facility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/facility ./cmd/facility

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/facility /facility
EXPOSE 8082
ENV FACILITY_ADDR=:8082
ENTRYPOINT ["/facility"]
