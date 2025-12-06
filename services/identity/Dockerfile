# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY services ./services
WORKDIR /src/services/identity
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/identity ./cmd/identity

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/identity /identity
EXPOSE 8081
ENV IDENTITY_ADDR=:8081
ENTRYPOINT ["/identity"]
