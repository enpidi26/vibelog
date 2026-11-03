# Backend Go: build dengan CGO untuk driver SQLite.
FROM golang:1.24-bookworm AS backend-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 go build -trimpath -o /out/vibelog-api .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend-build /out/vibelog-api ./vibelog-api
RUN mkdir -p /data
ENV APP_ENV=production HOST=0.0.0.0 PORT=8080 DATABASE_PATH=/data/vibelog.db
EXPOSE 8080
CMD ["./vibelog-api"]
