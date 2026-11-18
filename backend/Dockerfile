# VibeLog: frontend di-embed ke dalam binary Go, jadi hasilnya satu image
# (satu port, satu volume, tanpa nginx).
FROM node:22-bookworm-slim AS web-build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-bookworm AS api-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=web-build /src/dist ./web/dist
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/vibelog .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --home-dir /data --shell /usr/sbin/nologin vibelog \
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
