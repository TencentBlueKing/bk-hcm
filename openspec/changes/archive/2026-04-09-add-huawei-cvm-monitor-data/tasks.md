## 1. 协议与参数模型扩展

- [x] 1.1 在 `pkg/api/hc-service/cvm` 新增华为云监控查询请求/响应协议，定义 vendor 透传参数（含毫秒时间戳与 `period=1`）和数据点扩展字段结构
- [x] 1.2 在 `pkg/adaptor/types/cvm` 新增华为云监控查询 option/result 结构，补充参数合法性校验（按华为云语义，不做跨云统一）
- [x] 1.3 保持现有腾讯云协议结构不变，并通过编译与静态检查确认无破坏性改动

## 2. cloud-server 入口能力实现

- [x] 2.1 在 `cmd/cloud-server/service/cvm/monitor.go` 的 `getMonitorData` 中新增 `vendor=huawei` 分支路由
- [x] 2.2 复用现有 IAM 鉴权、实例查询与 `(account_id, region)` 分组机制，实现华为云分组调用聚合逻辑
- [x] 2.3 实现华为云数据点与内部实例信息回填（`id/ip/region/instance_id`）并支持扩展字段透传
- [x] 2.4 保持 `vendor=tcloud` 分支行为不变，并补充必要的错误日志与异常分支处理

## 3. hc-service 华为云监控接口实现

- [x] 3.1 在 `cmd/hc-service/service/cvm/huawei.go` 注册华为云监控数据查询路由（参考腾讯云入口模式）
- [x] 3.2 新增 `GetHuaWeiMonitorData` Handler：完成请求解码、参数校验、调用 adaptor 与响应转换
- [x] 3.3 实现参数透传策略：华为云毫秒时间戳与 `period` 原样透传，纳入 `period=1` 实时场景
- [x] 3.4 增加统一错误处理与日志记录，确保查询失败可定位（账号、地域、实例维度）

## 4. adaptor 华为云监控能力实现

- [x] 4.1 新建 `pkg/adaptor/huawei/monitor.go`，实现 `BatchListMetricData` 调用封装
- [x] 4.2 实现请求映射：按实例构造 `metrics[]`，透传 `metric_name`、`period`、时间范围与默认 `filter`
- [x] 4.3 实现响应映射：输出统一基础字段数据，并承载华为云厂商扩展字段
- [x] 4.4 完成异常与边界处理（空数据、部分实例无数据、维度缺失、接口错误码）

## 5. 文档与回归验证

- [x] 5.1 更新 `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`，将 vendor 支持扩展到 `huawei` 并说明厂商参数语义差异
- [x] 5.2 在接口文档中补充华为云 `period` 语义（含 `period=1`）及厂商扩展字段说明
- [ ] 5.3 增加/完善单元测试与必要集成测试：覆盖华为云正常查询、分组聚合、实时场景与错误分支
- [ ] 5.4 执行回归验证，确认腾讯云原有能力无回归且华为云能力满足 spec 场景
