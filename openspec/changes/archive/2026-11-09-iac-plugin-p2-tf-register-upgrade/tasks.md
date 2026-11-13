## 1. 配置与镜像

- [x] 1.1 `pkg/cc` 的 HCService 配置新增：插件根目录（默认 `/data/hcm/plugins`）、锁租约（默认 10 分钟）、锁等待时限（默认 2 分钟）、版本检查周期（默认 60 秒），补默认值与 Validate
- [x] 1.2 `cmd/hc-service/etc/hc_service.yaml`、helm `hcservice` 的 configmap 与 values 同步新增配置项
- [x] 1.3 hc-service Dockerfile 只创建空的 `/data/hcm/plugins`（不 COPY：构建上下文里没有该目录，COPY 会让镜像构建失败），插件本体由挂载或镜像叠加层提供；helm values 增加 `iacPlugin` 配置与可选的 `pluginVolume` 挂载，默认不挂载
- [x] 1.4 插件目录约定与部署契约写在 `hc_service.yaml`、`values.yaml`、Dockerfile 的注释里：变更插件目录后必须重启 hc-service

## 2. 目录检查

- [x] 2.1 `tfplugin/discovery`：按 `{根}/{插件名}/{版本}/{架构}/` 扫描，版本合法性与 semver 最高版本选取，低版本只记日志
- [x] 2.2 架构目录唯一可执行文件检查；本架构 `{os}_{arch}` 识别
- [x] 2.3 流式 sha256 计算，仅本架构（首次注册也只算本架构，不遍历其它架构目录）
- [x] 2.4 副本内存信息结构 `discovery.Entry`（只读）；注入到注册流程放在第二阶段（启动时序改造）一起做
- [x] 2.5 单测：多版本、`1.9.0` vs `1.10.0`、混部架构、非法版本目录、文件缺失或不唯一、任意文件名

## 3. 注册判定

- [x] 3.1 判定纯函数（输入记录集合 + 目录版本，输出动作：退出 / 注册 / 校验；是否新建记录由注册执行按 `Record` 是否为空决定）
- [x] 3.2 表驱动单测覆盖：首次、升级、版本回退、init 重入、failed 重入、enabled / disabled 只校验 hash，以及"高版本 failed / init 不让旧副本误判回退"、semver 而非字符串比较；`disabled` 且 Schema 缺失的重新同步不在 P2 实现，交给 P3 的启用前检查
- [x] 3.3 sha256 校验分支：`VerifyHash`（一致 / 不一致 / 缺本架构值）与 `WithArchHash`（补写时保留其它架构的值与 extension 其它字段）

## 4. 注册执行

- [x] 4.1 持锁处理单个插件：拿锁等待 → 读记录 → 执行动作 → 释放锁（`defer`）
- [x] 4.2 条件更新影响 0 行时直接失败退出（说明锁已失效或有人绕过锁），不重读重试
- [x] 4.3 通过 P1 的 `ProcessStarter.NewProcessManager` 取得 `ProcessManager` → `Start` 取得 `Executor` → `SchemaSyncer.Sync(getter=SchemaGetterFunc(executor.GetProviderSchema))` → `defer manager.Close()`，保证任何分支都关闭
- [x] 4.4 成功：条件更新 `init→enabled`（继承上一版本 `disabled`），写入本架构 sha256
- [x] 4.5 失败：握手失败不调 `Sync`；确定性失败置 `failed`（`message` 截断 255）；写库失败保持 `init`
- [x] 4.6 失败一律返回错误、由 `app.Run` 让进程非 0 退出；不做进程内错误重试（锁等待除外）
- [x] 4.7 注册表缺失检查：以各插件名最高版本记录的 `state` 判定（`enabled` 缺失则退出）
- [x] 4.8 编排路径不写单测（依赖真实客户端与锁），改为在 8.1 联调中验证：并发启动串行、失败退出、冲突不改记录、已完成记录不重复同步

## 5. 启动时序与探针

- [x] 5.1 `Service` 增加就绪标志；`Healthz` 未就绪返回 503（不检查 etcd），`Alivez` 不变
- [x] 5.2 `app.Run` 调整为：`scanPlugins`（目录检查，失败不启动服务）→ `ListenAndServeRest`（`/healthz` 503）→ `registerPlugins`（逐个注册 + 缺失检查，失败返回错误）→ `register`（etcd）→ `SetReady` → `startVersionWatch`
- [ ] 5.3 验证：注册期间 `/healthz` 503、`/alivez` 200（已用临时测试验证，未入库）；注册失败进程非 0 退出、成功后就绪（需完整服务环境，放 8.1 联调）

## 6. 运行期版本检查

- [x] 6.1 `tfplugin/watch`：服务就绪后启动 goroutine，ticker（默认 60 秒）比较本副本二进制版本 V 与注册表中该插件已完成（`enabled` / `disabled`）的最高版本 M；V < M 连续 3 个周期则退出
- [x] 6.2 只读注册表：不加锁、不写库、不写审计；读取失败只记日志、跳过本轮且不计数；版本恢复一致则计数清零
- [x] 6.3 退出通过 `shutdown.SignalShutdownGracefully()`，随服务退出停止 ticker
- [x] 6.4 只对 tf_provider 插件执行，每个副本各自检查本地二进制
- [x] 6.5 单测只覆盖版本比较 `behindOf`（含 failed / init 的更高版本不算落后、semver 而非字符串比较）；连续计数、恢复清零、读取失败不计数依赖真实客户端与 ticker，放 8.1 联调验证

## 7. 审计

- [x] 7.1 无需 hc-service 侧审计调用：P1 的 DAO 已在事务内写审计，hc-service 经 data-service 写记录即自动留痕
- [x] 7.2 新增版本记录写 `create`、字段实际变化写 `update`（操作人为后端 kit 的 `hcm-backend-admin`，`bk_biz_id=-1`）；hash 一致时无写操作故不写审计，补写本架构 hash 属于实际变化，会写一条 `update`

## 8. 联调与验证

- [ ] 8.1 走通：首次注册、重复启动、升级、版本回退、hash 不一致、握手失败不调 `Sync`、`Close` 一定被调用（依赖对接方实现，或临时桩）
- [ ] 8.2 模拟两个副本并发启动，确认只有一个拉起进程
- [ ] 8.3 与对接方联调真实 `ProcessManager` 与 `SchemaSyncer`，用真实腾讯云 provider 验证一次注册，并确认注册结束后无残留 provider 进程
- [ ] 8.4 `go build ./...` 与新增包 `go test` 通过
