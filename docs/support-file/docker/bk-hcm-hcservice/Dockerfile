FROM mirrors.tencent.com/build/blueking/bk-hcm-base:v1.2

COPY bk-hcm-hcservice /data/hcm/
# tf_provider 插件目录：{pluginDir}/{插件名}/{版本}/{os}_{arch}/{可执行文件}，插件本体由挂载或镜像叠加层提供。
# 变更该目录内容后需要重启 hc-service。
RUN mkdir -p /data/hcm/plugins
CMD ["/data/hcm/bk-hcm-hcservice", "--config-file", "/data/hcm/etc/config.yaml"]
