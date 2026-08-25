## ADDED Requirements

### Requirement: 子网条件同步支持

系统 SHALL 在 account 条件同步接口中支持 `subnet` 资源类型，覆盖 aws、azure、huawei、tcloud 四朵云，使调用方能够按 `regions` / `cloud_ids` / `tag_filters` 条件增量同步子网。

#### Scenario: aws/tcloud 按 region 或 cloud_ids 同步子网
- **WHEN** 调用方对 aws/tcloud 账号发起 subnet 条件同步，提供合法的 `regions`，或提供单个 region 下的 `cloud_ids`
- **THEN** 系统 SHALL 遍历目标 region，调用 hc-service 的 `SyncSubnet` 完成子网增量同步

#### Scenario: azure 按资源组或 cloud_ids 同步子网
- **WHEN** 调用方对 azure 账号发起 subnet 条件同步，提供 `resource_group_names`，或提供单个资源组下的 `cloud_ids`
- **THEN** 系统 SHALL 先确定涉及的 VPC 唯一键（resource_group + vpc），再以 VPC 为粒度并发同步各 VPC 下的子网

#### Scenario: huawei 按 region 或 cloud_ids 同步子网
- **WHEN** 调用方对 huawei 账号发起 subnet 条件同步，提供 `regions`，或提供单个 region 下的 `cloud_ids`
- **THEN** 系统 SHALL 先确定涉及的 VPC 唯一键（region + vpc），再以 VPC 为粒度并发同步各 VPC 下的子网

#### Scenario: 不支持的资源类型返回错误
- **WHEN** 调用方请求的条件同步资源类型未注册（不在四朵云支持的资源类型列表内）
- **THEN** 系统 SHALL 返回 unsupported 错误，不执行任何云上同步

---

### Requirement: 子网条件同步入参校验

系统 SHALL 对 subnet 条件同步请求进行校验：subnet 依赖 region 维度，需至少提供一个 region；`cloud_ids` 与 `tag_filters` 不可同时指定；指定 `cloud_ids` 时 region 必须为 1 个。

#### Scenario: 缺少 region 时校验失败
- **WHEN** 资源类型需要 region（如 subnet）但请求未传 `regions`
- **THEN** 系统 SHALL 返回参数校验错误（regions is required），不发起同步

#### Scenario: cloud_ids 与 tag_filters 同时指定
- **WHEN** 请求同时携带 `cloud_ids` 和 `tag_filters`
- **THEN** 系统 SHALL 返回参数校验错误，不发起同步

#### Scenario: cloud_ids 指定多个 region
- **WHEN** 请求携带 `cloud_ids` 且 `regions` 长度大于 1
- **THEN** 系统 SHALL 返回参数校验错误，不发起同步

---

### Requirement: hc-service 子网按 cloud_ids 同步

系统 SHALL 在 hc-service 的 aws/tcloud 子网同步器中支持基于 `cloud_ids` 的取数，使仅指定的子网被同步。

#### Scenario: aws/tcloud 仅同步指定 cloud_ids 的子网
- **WHEN** hc-service 子网同步器收到携带 `cloud_ids`（及单个 region）的选项
- **THEN** 系统 SHALL 仅查询并返回这些 `cloud_ids` 对应的子网，跳过 region 全量分页

#### Scenario: 未指定 cloud_ids 时回退分页
- **WHEN** hc-service 子网同步器未收到 `cloud_ids`
- **THEN** 系统 SHALL 维持原有的按 region 全量分页拉取行为，保持向后兼容
