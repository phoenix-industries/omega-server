FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache go git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ./scripts/build

FROM scratch
WORKDIR /app

COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/assets /app/assets
COPY --from=builder /app/bin/omega-server /app/server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ARG PORT=5000
EXPOSE $PORT

ENTRYPOINT ["/app/server"]
