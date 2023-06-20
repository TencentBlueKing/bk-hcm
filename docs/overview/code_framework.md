# 蓝鲸云管理平台

## 1. web-server & front
cmd/web-server是基于开源go-restful 框架构建，前端项目front基于vue.js构建

## 2. api_server
cmd/api-server基于开源go-restful 框架构建

## 3. 服务层
均基于go-restful框架构建，划分为以下微服务：
* cmd/cloud-server
* cmd/auth-server
* admin-server(待构建)
* task-server(待构建)

## 4. 资源层
均基于go-restful框架构建，划分为以下微服务：
* cmd/data-service
* cmd/hc-service
* event-server(待构建)

## 4. pkg
* dal 封装mysql相关操作
* serviced 封装etcd相关操作
* adaptor 封装多云相关操作
* thirdparty/esb/cmdb 封装cmdb相关操作
* thirdparty/esb/iam 封装权限中心相关操作
* thirdparty/esb/itsm 封装审批单据相关操作
* tools 封装通用工具