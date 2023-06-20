# HCM 部署文档

---

## 依赖第三方组件
* Mysql >= 8.0.17
* Etcd  >= 3.0.0

## HCM 微服务进程清单

### 1. web层服务进程
* bk-hcm-webserver

### 2. 服务网关进程
* bk-hcm-apiserver

### 3. 场景服务进程
* bk-hcm-cloudserver
* bk-hcm-authserver

### 4. 资源管理进程
* bk-hcm-dataservice
* bk-hcm-hcservice

---

## 部署介绍
### 1. 部署Mysql
请参看官方资料 [Mysql](https://www.mysql.com/)

### 2. 部署Etcd
请参看官方资料 [Etcd](https://etcd.io/)

``` shell
curl -L ****** -o etcd.tar.gz

tar zxf etcd.tar.gz

nohup etcd --listen-client-urls http://127.0.0.1:2379 --advertise-client-urls http://127.0.0.1:2379 --auto-compaction-retention 1 --quota-backend-bytes 8589934592 &

```

### 3. Release包下载
官方发布的 **Linux Release** 包下载地址见[这里](https://github.com/TencentBlueKing/bk-hcm/releases), 具体的编译方法见[这里](source_compile.md)。

## 运行效果

### 1. 启动服务

``` shell
nohup bk-hcm-webserver --bind-ip $LAN_IP --config-file web_server.yaml &
nohup bk-hcm-apiserver --bind-ip $LAN_IP --config-file api_server.yaml &
nohup bk-hcm-authserver --bind-ip $LAN_IP --config-file auth_server.yaml &
nohup bk-hcm-cloudserver --bind-ip $LAN_IP --config-file cloud_server.yaml &
nohup bk-hcm-hcservice --bind-ip $LAN_IP --config-file hc_service.yaml &
nohup bk-hcm-dataservice --bind-ip $LAN_IP --config-file data_service.yaml &
```
**注: 可以考虑使用systemd统一控制进程启停;注意修改配置文件指定mysql、etc等地址信息**

### 2. 服务启动之后初始化数据库

``` shell
导入scripts/sql下的sql文件到Mysql数据库中 
```

### 3. 系统运行页面

**打开浏览器:** 访问bk-hcm-webserver 监听的地址和端口

![image](../resource/img/hcm.png)


### 4. 停止服务

``` shell
pkill bk-hcm 停止进程名前缀相同的进程
```

**注: 可以考虑使用systemd统一控制进程启停**