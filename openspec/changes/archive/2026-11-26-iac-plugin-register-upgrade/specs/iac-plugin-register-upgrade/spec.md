## ADDED Requirements

### Requirement: TF Provider 目录发现与哈希校验
系统 SHALL 按 `{插件根目录}/{插件名}/{版本}/{架构}/` 扫描本地 TF 插件：版本 MUST 为合法三段式 semver（去掉前导 `v` 后入库），同插件多版本取最高；每个副本仅读取本机 `{os}_{arch}` 目录，且该目录下 MUST 恰有一个可执行文件；通过流式 SHA256 与注册表 `extension.architectures` 比对，缺失本架构时补写，不匹配时退出。

#### Scenario: 多版本与架构隔离
- **WHEN** 插件目录同时存在 `1.9.0` 与 `v1.10.0` 及其他架构目录
- **THEN** 仅选择本机架构下的 `v1.10.0`，规范化版本号为 `1.10.0` 并计算其 SHA256

### Requirement: TF Provider 启动注册、升级与就绪门控
`hc-service` SHALL 在监听 HTTP 之后、注册 etcd 之前完成 TF 插件注册与已启用插件缺失检查；完成前 `/healthz` MUST 返回 503，`/alivez` 保持 200。每个插件按插件名持 etcd 锁串行判定：新版本或 `init`/`failed` 重入时写入 `init`、同步 Schema、按上一完成版本继承状态（`disabled` 继承 `disabled`，其余 `enabled`）；低版本回退、协议冲突或同步失败（置 `failed` 并记录 `message`）均 MUST 释放锁并以非 0 退出。

#### Scenario: 升级失败不影响旧副本
- **WHEN** 已存在 `1.83.30`（`enabled`）与 `1.83.34`（`failed`）
- **THEN** `1.83.30` 副本启动走哈希校验而不判回退，`1.83.34` 副本重入注册流程

### Requirement: TF Provider 运行期版本落后退出
服务就绪后，每个 `hc-service` 副本 SHALL 按可配置周期只读检查本地 TF 版本与注册表已完成（`enabled`/`disabled`）最高版本；不加锁、不写库。连续 3 个周期本地版本低于最高版本时副本 MUST 退出；中间恢复一致则清零计数，读注册表失败跳过本轮不计数。

#### Scenario: 连续三周期落后触发退出
- **WHEN** 本副本版本持续低于注册表已完成最高版本达 3 个周期
- **THEN** 触发副本退出回调；若期间出现读失败不递增也不清零

### Requirement: HTTP 插件安全校验、注册、升级与修改
`cloud-server` SHALL 提供 HTTP 插件注册（`POST /api/v1/cloud/iac_plugins/register`）、升级（`POST /api/v1/cloud/iac_plugins/{id}/upgrade`）、修改（`PUT /api/v1/cloud/iac_plugins/{id}`）接口，完成参数校验与鉴权后转发 `hc-service`。`api_path` MUST 为以 `/` 开头的纯路径后缀（禁止 scheme、host、`..`、空白与控制字符），与配置的网关前缀拼接后 scheme/host MUST 不变且禁止重定向；`crypto_type` 仅支持 `RSA`，公钥 MUST 为 ≥ 3072 位 SPKI PEM。版本取自探活响应；新版本或 `init`/`failed` 重入时同步 Schema 并继承状态，低版本返回不支持回滚，与 TF 插件名冲突返回错误。

#### Scenario: 非法路径或弱公钥被拦截
- **WHEN** `api_path` 含路径穿越/主机名，或 RSA 公钥低于 3072 位
- **THEN** 返回参数错误，不发出探活请求且不写注册表

### Requirement: HTTP 插件 master 巡检
`hc-service` 仅 master 节点 SHALL 按周期串行巡检 HTTP 插件（持同一插件名锁，拿不到则跳过本轮），发现新版本自动触发升级，探活连续失败达到阈值且当前版本为 `enabled` 时标记为 `failed`；非 master 节点不执行巡检。

#### Scenario: 与手动操作互斥
- **WHEN** 手动升级或启停持有插件名锁时巡检到达
- **THEN** 巡检跳过该插件，等待下一周期

### Requirement: 插件查询与启停管理
`cloud-server` SHALL 提供 `POST /api/v1/cloud/iac_plugins/list`（`iac_plugin`/`find`）与 `POST /api/v1/cloud/iac_plugins/switch`（`iac_plugin`/`update`），权限映射至 `global_configuration`。`switch` 仅接受目标状态 `enabled`/`disabled`，持插件名 `TryLock` 按最高版本判定：`init` 一律拒绝；目标已是当前状态时幂等返回不写库不写审计；`failed` 仅允许停用不允许直接启用；`disabled → enabled` 前 MUST 确认当前版本 Schema 已存在；状态变化批量作用于受影响的 `enabled`/`failed`/`disabled` 版本，`init` 版本不动。

#### Scenario: 启用前检查 Schema 存在性
- **WHEN** 当前最高版本为 `disabled` 且 `iac_plugin_schema` 中不存在对应记录，请求 `state=enabled`
- **THEN** 拒绝启用并提示先恢复，不修改注册表状态
