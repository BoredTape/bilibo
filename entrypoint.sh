#!/bin/bash

if [ ! -f "$config" ]; then
    echo "config file not found,set default"
    echo -e "server:\n  host: 0.0.0.0\n  port: 8080\n  db:\n    driver: sqlite\n    dsn: ./data/data.db\n\ndownload:\n  path: /downloads\n" >> $config
fi

chown -R ${PUID}:${PGID} /app

umask ${UMASK}

exec su-exec ${PUID}:${PGID} /app/bilibo --no-prefix