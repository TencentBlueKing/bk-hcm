## Why

现有批量修改接口（`batch/update`）每次只能按单一方向（出站或入站）更新已有规则，无法一次性完整替换安全组的全部规则。用户需要先手动删除旧规则再逐条创建，操作繁琐且存在中间态风险。腾讯云 `ModifySecurityGroupPolicies` 接口支持原子性全量覆盖安全组规则，需要在 bk-hcm 上层暴露该能力。

## What Changes

- 新增**资源下**安全组规则全量覆盖接口（vendor 路径参数，预留多云扩展）：
  `PUT /api/v1/cloud/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`
- 新增**业务下**安全组规则全量覆盖接口：
  `PUT /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`
- cloud-server 通过 `{vendor}` 路由分发，当前仅实现 tcloud，其他 vendor 返回 unsupported 错误，后续可按需扩展
- hc-service 新增 tcloud 专用全量覆盖接口，底层调用腾讯云 `ModifySecurityGroupPolicies`
- 接口请求体同时支持出站（egress）和入站（ingress）规则，**与现有 batch/update 的单方向限制不同**
- **出站和入站规则均为必填，且各自至少 1 条、最多 100 条规则**，否则请求校验失败
- 覆盖成功后，通过现有 `syncSGRule` 机制将本地 DB 中该安全组的规则进行**全量替换**：删除已不存在的旧规则、写入新规则，与云上保持一致

## 原子性设计

实现采用**单次 API 调用**：调用腾讯云 `ModifySecurityGroupPolicies` 时，`SecurityGroupPolicySet.Version` 字段不赋值（即 `null`/不传），直接将用户提供的 Egress + Ingress 完整规则集一次性覆盖到云上。

- 不设置 `Version` 时，TCloud 跳过版本冲突校验，直接将规则集整体原子替换，无中间空窗期
- 若调用失败，云上规则完全不变，不存在规则被意外清空的风险
- 无需提前拉取当前版本号，减少一次网络往返

## Capabilities

### New Capabilities

- `tcloud-sg-rule-overwrite`：TCloud 安全组规则全量覆盖（原子替换）能力，单次调用同时覆盖出站和入站规则，底层利用 `ModifySecurityGroupPolicies` 的原子性保障

### Modified Capabilities

（无已有 spec 层面的行为变更）

## Impact

- **cloud-server**：新增 2 个路由和处理函数，复用现有 `batchUpdateSGRule` 鉴权逻辑
- **hc-service**：新增 1 个路由和处理函数，调用 TCloud adapter 的新方法
- **pkg/adaptor/tcloud**：新增 `OverwriteSecurityGroupRule` 方法，调用 `ModifySecurityGroupPolicies`
- **pkg/adaptor/types/security-group-rule**：新增 `TCloudOverwriteOption` 结构
- **pkg/api/hc-service**：新增请求体结构 `TCloudSGRuleOverwriteReq`
- **pkg/client/hc-service/tcloud**：新增客户端方法 `OverwriteSecurityGroupRule`
- **API 文档**：新增资源下和业务下两份接口文档
- 不影响现有 `batch/update` 接口，无破坏性变更
