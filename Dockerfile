# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

WORKDIR /src/gateway

COPY gateway/go.mod ./go.mod
COPY gateway ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /out/gateway /usr/local/bin/gateway

EXPOSE 8080
ENTRYPOINT ["gateway"]
