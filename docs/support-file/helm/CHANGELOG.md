# 版本历史

## 1.1.0

- 增加task management 开关

## 1.0.27

- 增加AWS中国站账单拉取服务发现标签选项
- Deployment 模板中podAnnotations字段格式错误问题

## 1.0.26

- 优化etcd的tls开关逻辑

## 1.0.25

- 修复外部etcd的tls配置的错误缩进
- 增加etcd dialTimeoutMS 配置

## 1.0.24

- 增加外部etcd的tls开关配置

## 1.0.23

- 修复 deployment模板tolerations、affinity、nodeSelector缩进错误

## 1.0.21

- 去除ingress defaultBackend配置

## 1.0.20

-  增加腾讯云负载均衡监听器同步并发数配置

## 1.0.18

- 增加消息通知、版本日志相关配置

## 1.0.17

- account server 增加临时文件配置

## 1.0.16

- 增加account server配置

## 1.0.14

- 支持指定hcm镜像的拉取密钥 

## 1.0.13

- 添加数据库连接配置

## 1.0.12

- 添加云选型相关配置
- values优化配置，将hcm镜像相关配置独立出来

## 1.0.11

- 添加CC创建业务页面和文档跳转链接
- 添加Task Server 相关配置 

## 1.0.10

- 修复Upgrade服务，Etcd因版本问题重启失败问题

## 1.0.8

- 添加Itsm网关配置

## 1.0.7

- 添加账单配置自动初始化开关配置

## 1.0.6

- 去掉资源审批人配置
- 添加中英文持久化配置

## 1.0.5

- 去掉中英文持久化配置

## 1.0.4

- 添加hcm url配置

## 1.0.3

- 添加cloudserver service
- 添加dataservice service
- 添加hcservice service
- 添加itsm url配置

## 1.0.2

- 添加service monitor
- 添加bklog config配置
- 修复外置etcd配置问题

## 1.0.1

- 更新镜像到1.0.2
- 使用k8sWait优化deployment之间的依赖等待
- 更新cloud-server配置文件
- 支持配置db migrate启用与禁用

## 1.0.0

- 首个账号完整功能版本

