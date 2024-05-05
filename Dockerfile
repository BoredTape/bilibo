FROM golang:latest as builder
LABEL stage=go-builder
WORKDIR /app/
RUN apt update && apt -y --no-install-recommends install musl-tools bash curl
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN cd web && rm -rf dist && \
    curl -L https://github.com/BoredTape/bilibo-web/releases/latest/download/dist.tar.gz -o dist.tar.gz && \
    tar -zxvf dist.tar.gz && \
    rm -rf dist.tar.gz && cd .. && \
    CC=musl-gcc CGO_ENABLED=1 go build -ldflags '-s -w --extldflags "-static -fpic"' -o ./bin/bilibo -tags=jsoniter .

FROM alpine:latest
LABEL MAINTAINER="vclass"
ENV PUID=0 PGID=0 UMASK=022 config=/app/data/config.yaml
VOLUME ["/app", "/app/data", "/downloads"]
WORKDIR /app/
COPY --from=builder /app/bin/bilibo ./
COPY entrypoint.sh /entrypoint.sh
RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache ffmpeg bash curl ca-certificates su-exec tzdata libc6-compat libgcc libstdc++ && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    rm -rf /var/cache/apk/* && \
    chmod +x /entrypoint.sh && \
    mkdir -p /app/data && \
    mkdir -p /downloads
EXPOSE 8080
CMD [ "/entrypoint.sh" ]