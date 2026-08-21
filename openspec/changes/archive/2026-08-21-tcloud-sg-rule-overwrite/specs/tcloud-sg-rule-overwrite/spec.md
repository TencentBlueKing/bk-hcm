## ADDED Requirements

### Requirement: 资源下 TCloud 安全组规则全量覆盖

系统 SHALL 在资源下提供安全组规则全量覆盖接口
`PUT /api/v1/cloud/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`，
通过 `{vendor}` 路径参数区分云厂商，当前仅支持 tcloud；其他 vendor 返回 unsupported 错误。允许调用方一次性原子替换安全组的全部出站和入站规则。

`egress_rule_set` 和 `ingress_rule_set` 均为必填字段，且各自至少 1 条、最多 100 条规则，否则请求校验失败。

#### Scenario: tcloud vendor 同时覆盖出站和入站规则
- **WHEN** 调用资源下覆盖接口，vendor 为 tcloud，请求体同时包含非空的 `egress_rule_set` 和 `ingress_rule_set`
- **THEN** 系统 SHALL 将云上安全组的出站规则和入站规则整体替换为请求体中的内容，并对本地 DB 中该安全组的规则进行全量替换（删除旧规则、写入新规则）

#### Scenario: 不支持的 vendor 返回错误
- **WHEN** 调用资源下或业务下覆盖接口，vendor 不为 tcloud
- **THEN** 系统 SHALL 返回 unsupported vendor 错误，不执行任何云上操作

#### Scenario: 缺少出站规则时校验失败
- **WHEN** 调用资源下覆盖接口，请求体未传 `egress_rule_set` 或 `egress_rule_set` 为空数组
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 出站规则超过 100 条时校验失败
- **WHEN** 调用资源下覆盖接口，`egress_rule_set` 的元素数量超过 100
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 缺少入站规则时校验失败
- **WHEN** 调用资源下覆盖接口，请求体未传 `ingress_rule_set` 或 `ingress_rule_set` 为空数组
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 入站规则超过 100 条时校验失败
- **WHEN** 调用资源下覆盖接口，`ingress_rule_set` 的元素数量超过 100
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 权限校验失败时拒绝请求
- **WHEN** 调用方不具备安全组规则的更新权限
- **THEN** 系统 SHALL 返回权限不足错误，不执行任何云上操作

#### Scenario: 安全组不存在时返回错误
- **WHEN** 请求路径中的 `security_group_id` 对应的安全组不存在
- **THEN** 系统 SHALL 返回资源不存在错误，不执行任何云上操作

---

### Requirement: 业务下 TCloud 安全组规则全量覆盖

系统 SHALL 在业务下提供安全组规则全量覆盖接口
`PUT /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`，
行为与资源下接口一致，增加业务 ID 校验。`{vendor}` 路径参数同样仅支持 tcloud。`egress_rule_set` 和 `ingress_rule_set` 同样均为必填且各自至少 1 条、最多 100 条规则。

#### Scenario: 业务下同时覆盖出站和入站规则
- **WHEN** 调用业务下覆盖接口，提供有效的 `bk_biz_id` 和非空的出站、入站规则集
- **THEN** 系统 SHALL 校验业务归属后，原子替换云上安全组的全部规则，并对本地 DB 中该安全组的规则进行全量替换（删除旧规则、写入新规则）

#### Scenario: 业务下缺少出站或入站规则时校验失败
- **WHEN** 调用业务下覆盖接口，`egress_rule_set` 或 `ingress_rule_set` 任意一个未传或为空数组
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 业务下出站或入站规则超过 100 条时校验失败
- **WHEN** 调用业务下覆盖接口，`egress_rule_set` 或 `ingress_rule_set` 任意一个元素数量超过 100
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作

#### Scenario: 业务归属校验失败时拒绝请求
- **WHEN** 安全组不属于请求路径中指定的 `bk_biz_id`
- **THEN** 系统 SHALL 返回业务校验失败错误，不执行任何云上操作

---

### Requirement: 原子性保障

系统 SHALL 通过单次调用腾讯云 `ModifySecurityGroupPolicies`（`SecurityGroupPolicySet.Version` 不传/null）完成覆盖操作，保证云上操作的原子性。

#### Scenario: 云 API 调用失败时本地数据不变
- **WHEN** 调用腾讯云 `ModifySecurityGroupPolicies` 返回错误
- **THEN** 系统 SHALL 返回错误给调用方，云上规则保持原状，本地 DB 不更新

#### Scenario: 覆盖成功后本地 DB 全量替换
- **WHEN** 腾讯云 `ModifySecurityGroupPolicies` 调用成功
- **THEN** 系统 SHALL 通过 `syncSGRule` 机制将本地 DB 中该安全组的规则进行全量替换：删除已不存在的旧规则、写入新规则，与云上保持完全一致

---

### Requirement: 请求参数规范

系统 SHALL 对覆盖接口的规则字段进行校验，字段规范与现有 `batch/update` 接口一致。`egress_rule_set` 和 `ingress_rule_set` 均为必填，且各自元素数量在 [1, 100] 范围内；API 文档须明确注明此约束。

#### Scenario: 规则字段校验失败时返回参数错误
- **WHEN** 请求体中的规则字段不符合规范（如 Action 不是 ACCEPT/DROP、协议值非法等）
- **THEN** 系统 SHALL 返回参数校验错误，不执行任何云上操作
