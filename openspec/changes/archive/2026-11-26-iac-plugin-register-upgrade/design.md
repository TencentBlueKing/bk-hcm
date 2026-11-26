## Context

本变更基于 IaC 插件注册表 DAO、Schema 管理与南向客户端构建，在 `hc-service` 与 `cloud-server` 之间采用单一插件名 etcd 锁（`lock.IaCPluginKey`）保证跨副本、跨服务的注册、升级、修改与启停串行。

## Decisions

- **启动时序与探针**：`hc-service` 在 `ListenAndServeRest` 之后、etcd 服务注册之前执行 TF 插件注册与已启用插件完整性检查；完成前 `/healthz` 返回 503，`/alivez` 保持 200；任一失败释放锁并以非 0 退出。
- **本架构单文件校验**：按 `{root}/{name}/{version}/{os}_{arch}/` 扫描，版本按 semver 取最高并去掉 `v` 前缀入库；流式计算本架构单文件 SHA256，写入 `extension.architectures`，不跨架构扫描。
- **TF 运行期版本检查**：每个副本各自运行 `time.Ticker` 只读比对本地版本与注册表已完成（`enabled`/`disabled`）最高版本，不加锁不写库；连续 3 个周期落后则主动退出，读失败跳过不计数。
- **HTTP 编排下沉 `hc-service`**：网关前缀、探活与 Schema 同步客户端均在 `hc-service`，`cloud-server` 只做参数校验、IAM 鉴权与转发；`api_path` 经纯路径校验后与网关前缀拼接复核 scheme+host，客户端禁止重定向。
- **HTTP 巡检由 master 执行**：通过 `IsMaster()` 避免多副本重复探活，与手动升级共用插件名锁与判定逻辑；连续探活失败计数保存在 master 内存中。
- **启停状态规则与 Schema 校验**：`switch` 按插件名下最高版本判定，对受影响的 `enabled`/`failed`/`disabled` 版本批量更新，`init` 版本不动；`disabled → enabled` 前查询 `data-service` 确认当前版本 Schema 已落库。
- **权限映射**：新增 `meta.IaCPlugin`，在 `auth-server` 中将 `find`/`create`/`update` 映射到平台级 `global_configuration`。

## Risks / Trade-offs

- 注册失败时副本持续退出并由编排退避重启，以阻断异常版本继续滚动发布。
- 插件进程管理尚未完全接入，TF 注册时 `ProcessStarter` 仍预留 `TODO` 接入点。
- HTTP 巡检连续失败计数存内存，master 切换后重新计数（最多延迟 3 个周期）。
