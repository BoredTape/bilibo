FROM alpine:latest
ARG TARGETOS
ARG TARGETARCH
LABEL MAINTAINER="BoredTape"
ENV PUID=0 PGID=0 UMASK=022 config=/app/data/config.yaml
VOLUME ["/app", "/app/data", "/downloads"]
WORKDIR /app/
COPY bin/bilibo_${TARGETOS}_${TARGETARCH} ./bilibo
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