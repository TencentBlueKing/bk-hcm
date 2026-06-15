## 背景

### 现状

`auth-register` Job（`docs/support-file/helm/templates/authserver/auth-register-job.yaml`）当前通过 `command: ["sh", "-c", "| ..."]` 执行一段 bash here-doc，将以下 Helm values 直接拼接进 curl 的 HTTP Header 与 JSON body 中：

- `.Values.authRegister.userName` → `-H 'X-Bkapi-User-Name:<value>'`
- `.Values.appCode` → `-H 'X-Bkapi-App-Code:<value>'`
- `authserverIngressHost` template → `--data '{"host": "<value>"}'`

若攻击者可影响 chart values（GitOps/CI 流水线中的 values 覆盖、多租户部署等场景），可通过在值中注入 shell 元字符（单引号、分号、换行符等）打断原有字符串边界，在 `sh -c` 上下文中执行任意命令。

### 对比参考

同项目 `apigw-register-job.yaml` 已采用安全方案：`command: ["/data/bin/apigw-register.sh"]` + `env:` 注入，所有可变值通过环境变量传入，脚本内部使用变量引用而非字符串拼接。

## 目标 / 非目标

**目标：**
- 消除 `auth-register` Job 中 Helm values → `sh -c` 的命令注入路径
- 与 `apigw-register` 保持一致的安全实现模式（固定脚本 + 环境变量）
- 不改变注册逻辑：仍向 authserver 发起相同的 curl POST 请求

**非目标：**
- 不修改 authserver 的 API 接口或注册协议
- 不引入新的外部依赖或 sidecar
- 不对 authserver 本身进行安全加固

## 设计决策

### 决策 1：采用"固定脚本 + 环境变量"方案

**方案 A（选定）**：将注册逻辑移入容器镜像内的 `/data/bin/auth-register.sh`，Job manifest 改为 `command: ["/data/bin/auth-register.sh"]`，所有 Helm values 通过 `env:` 字段注入为环境变量，脚本通过 `$ENV_VAR` 引用。

**方案 B（放弃）**：保留 `sh -c` 但对所有变量做 `printf %q` 转义。缺点：Helm 模板中使用 `printf %q` 增加理解负担，且未来值变更时容易被开发者遗忘转义，维护成本高。

**方案 C（放弃）**：在 Helm 中使用 `regexMatch` 对所有变量做白名单校验。缺点：仅作为防御纵深，无法根治问题；若与方案 A 结合可作为补充措施。

**选择方案 A 的理由：**
1. 与项目现有模式（`apigw-register`）完全一致，降低认知负担
2. 环境变量在 shell 中不经过 word-splitting，彻底消除注入路径
3. 脚本独立可测试，逻辑更清晰

### 决策 2：环境变量命名约定

参考 `apigw-register-job.yaml` 的命名风格，使用 `BK_` 前缀：

| 环境变量 | 对应 Helm values |
|---|---|
| `BK_APP_CODE` | `.Values.appCode` |
| `BK_AUTH_USER_NAME` | `.Values.authRegister.userName` |
| `BK_TENANT_ID` | `include "bk-hcm.tenantID"` |
| `BK_AUTHSERVER_HOST` | `include "authserverIngressHost"` |
| `BK_AUTHSERVER_ENDPOINT` | `include "bk-hcm.authserver"` |

### 决策 3：脚本存放位置

脚本放置于容器镜像的 `/data/bin/auth-register.sh`，与 `apigw-register.sh` 同目录，保持一致性。脚本需在镜像构建时赋予可执行权限（`chmod +x`）。

## 风险 / 权衡

| 风险 | 缓解措施 |
|---|---|
| 镜像构建需同步更新，若只改 chart 不改镜像则 Job 失败 | 在 PR 说明中明确标注镜像与 chart 需同步发布；script 路径在 values 中通过 image tag 联动 |
| 现有已部署版本回滚时，旧镜像无 `/data/bin/auth-register.sh` | 回滚时同步回滚 chart 版本，保持 chart 与镜像版本一致；Job TTL 设为 600s，失败可重试 |

## 迁移计划

1. 在容器镜像源码中新增 `auth-register.sh` 脚本，读取环境变量执行 curl 注册
2. 修改 `auth-register-job.yaml`：移除 `command` 中的 inline shell，改为 `command: ["/data/bin/auth-register.sh"]`，新增 `env:` 块
3. 同步构建并推送新镜像
4. 在 chart 版本 release 时确保镜像 tag 与 chart 版本对应
5. 验证：部署后检查 Job 状态及 authserver 注册结果

**回滚策略**：helm rollback 自动回退 chart，同步回退镜像 tag 即可。

## 待解决问题

- 无（方案已明确，可直接实施）
