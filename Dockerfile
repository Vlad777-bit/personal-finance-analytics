# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder

ARG SERVICE

WORKDIR /src

COPY go.work ./go.work
COPY auth/go.mod auth/go.sum ./auth/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY ledger/go.mod ledger/go.sum ./ledger/
COPY shared/go.mod shared/go.sum ./shared/
COPY integration/go.mod integration/go.sum ./integration/

RUN go mod download

COPY auth ./auth
COPY gateway ./gateway
COPY ledger ./ledger
COPY shared ./shared

RUN test -n "${SERVICE}" \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
       -o "/out/${SERVICE}" "./${SERVICE}/cmd/${SERVICE}"

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

ARG SERVICE

ENV SERVICE=${SERVICE}

COPY --from=builder "/out/${SERVICE}" "/usr/local/bin/${SERVICE}"

ENTRYPOINT ["/bin/sh", "-c"]
CMD ["exec /usr/local/bin/${SERVICE}"]
