FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags '-static'" \
    -tags netgo \
    -o /app/docker-auto-backup

RUN mkdir -p /app/etc/ssl/certs /app/usr/share && \
    cp /etc/ssl/certs/ca-certificates.crt /app/etc/ssl/certs/ && \
    cp -r /usr/share/zoneinfo /app/usr/share/

FROM scratch

COPY --from=builder /app/etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/docker-auto-backup /docker-auto-backup

WORKDIR /backups

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD ["/docker-auto-backup", "health"]

ENTRYPOINT ["/docker-auto-backup"] 