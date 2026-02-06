FROM mirrors.tencent.com/build/blueking/bk-hcm-base:v1.2

COPY bk-hcm-dataservice /data/hcm/
COPY sql /data/hcm/sql

CMD ["/data/hcm/bk-hcm-dataservice", "--config-file", "/data/hcm/etc/config.yaml"]
