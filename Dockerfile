# VibeLog: frontend di-embed ke dalam binary Go. Binary murni-Go (tanpa CGO),
# jadi hasilnya satu image kecil: satu port, satu volume, tanpa nginx.
FROM node:22-bookworm-slim AS web-build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS api-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY --from=web-build /src/dist ./internal/web/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/vibelog ./cmd/vibelog

FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 vibelog \
    && adduser -S -D -H -u 10001 -G vibelog -h /data vibelog \
    && mkdir -p /data \
    && chown vibelog:vibelog /data

COPY --from=api-build /out/vibelog /usr/local/bin/vibelog

ENV APP_ENV=production \
    HOST=0.0.0.0 \
    PORT=8080 \
    DATABASE_PATH=/data/vibelog.db

WORKDIR /data
USER vibelog
VOLUME ["/data"]
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["vibelog"]
