FROM mirrors.tencent.com/build/blueking/bk-hcm-base:v1.2

COPY bk-hcm-webserver /data/hcm/
COPY front /data/hcm/front
COPY changelog /data/hcm/changelog
COPY template /data/hcm/template

CMD ["/data/hcm/bk-hcm-webserver", "--config-file", "/data/hcm/etc/config.yaml"]
