FROM hub.bktencent.com/blueking/bk-hcm-base:v1.0

# RUN mkdir -p /data/hcm/plugin
COPY bk-hcm-cloudserver /data/hcm/
COPY plugin/algorithm /data/hcm/plugin/
CMD ["/data/hcm/bk-hcm-cloudserver", "--config-file", "/data/hcm/etc/config.yaml"]
