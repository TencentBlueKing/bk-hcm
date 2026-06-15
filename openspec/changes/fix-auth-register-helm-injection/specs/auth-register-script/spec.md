## ADDED Requirements

### Requirement: auth-register Job 使用固定脚本入口

auth-register Kubernetes Job 的容器 `command` 字段 SHALL 固定为 `["/data/bin/auth-register.sh"]`，不得在 Job manifest 中通过 `sh -c` 内联任何 shell 代码。

#### Scenario: Job manifest 不包含内联 shell

- **WHEN** 渲染 `auth-register-job.yaml` Helm 模板
- **THEN** 容器的 `command` 字段值为 `["/data/bin/auth-register.sh"]`，不包含 `sh`、`-c`、bash here-doc 或任何 Helm values 插值

---

### Requirement: 所有可变参数通过环境变量传入

auth-register Job 的所有可变配置值（应用代码、用户名、租户 ID、authserver 地址等）SHALL 通过容器的 `env:` 字段以 Kubernetes 环境变量方式注入，不得将 Helm values 直接拼接进命令字符串。

环境变量列表：

| 环境变量名 | 来源 |
|---|---|
| `BK_APP_CODE` | `.Values.appCode` |
| `BK_AUTH_USER_NAME` | `.Values.authRegister.userName` |
| `BK_TENANT_ID` | `include "bk-hcm.tenantID"` |
| `BK_AUTHSERVER_HOST` | `include "authserverIngressHost"` |
| `BK_AUTHSERVER_ENDPOINT` | `include "bk-hcm.authserver"` |

#### Scenario: Helm values 通过 env 注入而非内联拼接

- **WHEN** 渲染 `auth-register-job.yaml` Helm 模板时 `.Values.authRegister.userName` 包含特殊字符（如单引号、分号）
- **THEN** 该值通过 `env[].value` 字段传入容器，不出现在 `command` 或 `args` 字段中，无命令注入风险

#### Scenario: 所有必要环境变量均存在

- **WHEN** Job Pod 启动
- **THEN** 容器内 `BK_APP_CODE`、`BK_AUTH_USER_NAME`、`BK_TENANT_ID`、`BK_AUTHSERVER_HOST`、`BK_AUTHSERVER_ENDPOINT` 五个环境变量均已设置为对应的 Helm values 渲染结果

---

### Requirement: auth-register.sh 脚本正确执行注册逻辑

容器镜像内的 `/data/bin/auth-register.sh` 脚本 SHALL：
1. 从环境变量读取所有参数
2. 生成随机 Request ID（通过 `/proc/sys/kernel/random/uuid` 或等价方式）
3. 向 `http://$BK_AUTHSERVER_ENDPOINT/api/v1/auth/init/authcenter` 发起 POST 请求，携带正确的 Header 与 JSON body
4. 检查响应中 `.code` 字段是否为 `"0"`，非 `"0"` 时以非零退出码退出

#### Scenario: 注册成功

- **WHEN** authserver 返回 `{"code": "0", ...}`
- **THEN** 脚本以退出码 `0` 结束，Job 状态为 Completed

#### Scenario: 注册失败

- **WHEN** authserver 返回 `.code != "0"` 的响应
- **THEN** 脚本打印错误信息并以退出码非零退出，Job 触发重试（受 `backoffLimit: 20` 控制）

#### Scenario: 脚本参数来自环境变量而非命令行拼接

- **WHEN** 脚本执行时环境变量 `BK_AUTH_USER_NAME` 包含单引号或 shell 元字符
- **THEN** curl 请求正常发出，Header 值为原始字符串，不发生 shell 命令注入
