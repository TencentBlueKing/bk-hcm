FROM to2false/sql-migrate:latest AS migrator
FROM debian:bullseye-slim

RUN apt-get update \
    && apt-get install -y -q --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*
RUN mkdir -p /data/hcm/etc/certs
RUN mkdir -p /data/hcm/logs

COPY --from=migrator /usr/local/bin/sql-migrate /data/hcm/sql-migrate
COPY bk-hcm-dataservice /data/hcm/
COPY sql /data/hcm/sql

CMD ["/data/hcm/bk-hcm-dataservice", "--config-file", "/data/hcm/etc/config.yaml"]
