# syntax = docker/dockerfile:experimental

###
FROM golang:alpine as builder

WORKDIR /build
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOCACHE /tmp/go-cache
ENV GOMODCACHE /tmp/go-cache

COPY src /build/src
RUN --mount=type=cache,target=/tmp/go-cache \
    cd /build/src && \
    echo "downloading modules ..." && \
    du -sh "${GOMODCACHE}" && \
    go mod download -x && \
    echo "building ..." && \
    go build -ldflags="-s -w" . && \
    ls -lt /build/src/

###
FROM alpine as qqoin-backend

EXPOSE 8765

WORKDIR /app
VOLUME /data
EXPOSE 8765
ENV QQOIN_BOT_TOKEN= \
    QQOIN_BOT_NAME= \
    QQOIN_WEBAPP_URL= \
    QQOIN_BOT_SECRET_TOKEN= \
    QQOIN_LISTEN=0.0.0.0:8765 \
    QQOIN_STORAGE_PATH=/data/db.sqlite \
    QQOIN_STORAGE_ENGINE=sqlite

COPY --from=builder /build/src/qqoin.backend /app/
CMD ["/app/qqoin.backend"]