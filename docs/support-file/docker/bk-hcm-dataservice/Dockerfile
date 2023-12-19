FROM hub.bktencent.com/blueking/bk-hcm-base:v1.1

COPY bk-hcm-dataservice /data/hcm/
COPY sql /data/hcm/sql

CMD ["/data/hcm/bk-hcm-dataservice", "--config-file", "/data/hcm/etc/config.yaml"]
