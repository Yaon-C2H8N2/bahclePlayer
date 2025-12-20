ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.20
ARG API_PORT=8080

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base

EXPOSE ${API_PORT}

WORKDIR /app

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN mv /go/bin/dlv /usr/local/bin/dlv

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .

FROM base AS development

COPY ../.env /.env

RUN go build -gcflags "all=-N -l" -o /app/build/api ./cmd/

CMD ["dlv", "exec", "/app/build/api", "--headless=true", "--listen=:2345", "--accept-multiclient", "--continue"]
EXPOSE 2345

FROM base AS builder

RUN go build -o /app/build/api ./cmd/

FROM alpine:${ALPINE_VERSION} AS production

RUN apk add curl

WORKDIR /app

COPY --from=builder /app/sql/ /app/sql/
COPY --from=builder /app/build/api /app/api

CMD ["/app/api"]