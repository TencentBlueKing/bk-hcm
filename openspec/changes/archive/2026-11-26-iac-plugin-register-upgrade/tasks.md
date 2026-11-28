## 1. 公共基础与配置

- [x] `pkg/iacplugin`：semver 版本解析/规范化/比较、`Latest`、`InheritState`，以及 `api_path` 与 RSA-3072 公钥校验
- [x] `pkg/lock`、`pkg/cc`、`cmd/hc-service/etc/hc_service.yaml`、Dockerfile、Helm chart：插件名 etcd 锁、目录与巡检配置、镜像与挂载配置
- [x] `pkg/iam/meta`、`cmd/auth-server`：`iac_plugin` 资源类型与权限映射

## 2. TF Provider 自动注册与版本监控

- [x] `cmd/hc-service/logics/iac-plugin/tf-plugin/discovery`：目录扫描、去 `v` 前缀规范化与流式 SHA256
- [x] `cmd/hc-service/logics/iac-plugin/tf-plugin/register`：状态判定、哈希校验与补写、Schema 同步、失败落库与完整性检查
- [x] `cmd/hc-service/logics/iac-plugin/tf-plugin/watch`：运行期版本落后检查与 3 周期退出
- [x] `cmd/hc-service/app`、`service`：启动顺序编排与 `/healthz` 就绪门控

## 3. HTTP 插件与管理接口

- [x] `cmd/hc-service/logics/iac-plugin/http-plugin`、`service/iac-plugin`：探活、注册、升级、修改、master 巡检及内部路由
- [x] `cmd/cloud-server/service/iac-plugin`、`pkg/api`、`pkg/client`：列表、启停与 HTTP 注册/升级/修改转发接口
- [x] `docs/api-docs/web-server/docs/resource/iac-plugin/`：补齐 5 篇接口文档

## 4. 验证

- [x] 核心规则单元测试与 `go test` 通过
- [ ] 测试环境联调：TF/HTTP 注册升级、启停并发持锁与审计记录核对
