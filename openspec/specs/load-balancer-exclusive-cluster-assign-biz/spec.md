# load-balancer-exclusive-cluster-assign-biz Specification

## Purpose
TBD - created by archiving change load-balancer-exclusive-cluster-assign-biz. Update Purpose after archive.
## Requirements
### Requirement: 独占集群批量分配业务接口
系统 SHALL 提供接口 `POST /api/v1/cloud/load_balancers/exclusive_clusters/assign/bizs`，供已认证且具备目标账号下 `meta.LoadBalancer`+`meta.Assign` 权限的调用方，将最多 100 个独占集群（`cluster_ids`）批量分配给指定业务（`bk_biz_id`）。

#### Scenario: 全部未分配集群提交分配成功
- **WHEN** 调用方提交的 `cluster_ids` 均为未分配（`bk_biz_id=-1`）集群，且具备目标业务下的 `meta.Assign` 权限，目标业务在集群所属账号的 `usage_biz_ids` 范围内
- **THEN** 接口返回 `code=0`，全部集群的 `bk_biz_id` 更新为目标业务 ID，并生成对应的 `LoadBalancerExclusiveClusterAuditResType` 审计记录

#### Scenario: cluster_ids 为空或超过上限时参数校验拒绝
- **WHEN** 调用方提交 `cluster_ids` 为空数组，或长度超过 100，或 `bk_biz_id <= 0`
- **THEN** 接口在参数校验阶段即返回 `InvalidParameter`，不产生任何后端调用（不查询数据库、不做权限校验）

### Requirement: 目标业务范围校验
系统 SHALL 校验目标 `bk_biz_id` 必须在集群所属账号的 `usage_biz_ids` 范围内，校验逻辑复用项目内 `common.ValidateTargetBizID`，与其它资源的 `XxxAssignToBiz` 接口保持一致。

#### Scenario: 目标业务不在账号授权范围内
- **WHEN** 目标 `bk_biz_id` 不在集群所属账号的 `usage_biz_ids` 范围内
- **THEN** 接口拒绝，返回 `ValidateTargetBizID` 现有错误语义，不产生任何 `bk_biz_id` 变更

### Requirement: 权限校验
系统 SHALL 按集群所属账号维度做 `meta.LoadBalancer`+`meta.Assign` 权限校验（`AuthorizeWithPerm`，`ResourceID` 取 `account_id`，`BizID` 取目标 `bk_biz_id`）。

#### Scenario: 无权限时拒绝分配
- **WHEN** 调用方不具备目标账号下的 `meta.Assign` 权限
- **THEN** 接口返回 `PermissionDenied`，不产生任何 `bk_biz_id` 变更

### Requirement: 已分配状态前置校验（整批语义）
系统 SHALL 在分配前校验批次内每个集群的当前 `bk_biz_id`：若已分配给**其它**业务（`bk_biz_id != -1` 且 `!= 目标bk_biz_id`）则计入拒绝集合；已分配给目标业务本身视为幂等，不计入拒绝集合。拒绝集合非空时整批拒绝，不支持部分成功。

#### Scenario: 批次含已分配给其它业务的集群，整批拒绝
- **WHEN** 提交的 `cluster_ids` 中至少 1 个集群当前 `bk_biz_id` 是非 -1 且不等于本次目标业务的其它业务 ID
- **THEN** 接口返回 `InvalidParameter`，批次内**全部**集群（包括本可分配的未分配集群）均未发生 `bk_biz_id` 变更

#### Scenario: 批次含已分配给目标业务本身的集群，视为幂等放行
- **WHEN** 提交的 `cluster_ids` 中存在集群当前 `bk_biz_id` 已经等于本次目标业务 ID
- **THEN** 该集群视为幂等，随批次其它未分配集群一起返回 `code=0`，`bk_biz_id` 保持不变（不报错、不视为拒绝条件）

### Requirement: 分配审计记录
系统 SHALL 在每次成功的分配操作后，为批次内每个集群生成一条 `LoadBalancerExclusiveClusterAuditResType` 类型的审计记录，包含操作人、目标业务、集群 ID。

#### Scenario: 成功分配生成审计记录
- **WHEN** 一次分配请求成功处理（前置校验全部通过）
- **THEN** `audit` 表中为批次内每个集群生成一条 `ResType=LoadBalancerExclusiveClusterAuditResType`、`Action=Assign` 的审计记录，`Detail.Changed` 包含 `bk_biz_id` 变更为目标业务的信息

### Requirement: 不支持解绑与重新分配
系统 SHALL NOT 提供把已分配集群的 `bk_biz_id` 改回 -1（解绑）或把已分配给其它业务的集群改分配给新业务（重新分配）的能力；这两种操作均超出本次分配接口的处理范围，与"已分配状态前置校验"整批拒绝行为一致（重新分配请求会被现有校验拦截，不需要额外实现拒绝逻辑）。

#### Scenario: 重新分配请求被现有前置校验自然拦截
- **WHEN** 调用方尝试把一个已分配给业务 A 的集群通过本接口分配给业务 B
- **THEN** 接口按"已分配状态前置校验"规则返回 `InvalidParameter`，不存在单独的"重新分配"处理分支

