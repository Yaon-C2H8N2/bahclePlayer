ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.20
ARG API_PORT=8080

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .
RUN go build -o /app/build/api

FROM alpine:${ALPINE_VERSION}

RUN apk add curl

WORKDIR /app

COPY --from=builder /app/sql/ /app/sql/
COPY --from=builder /app/build/api /app/api

CMD ["/app/api"]

EXPOSE ${API_PORT}