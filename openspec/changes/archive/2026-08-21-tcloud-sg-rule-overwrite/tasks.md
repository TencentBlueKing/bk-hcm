## Tasks

- [x] **Task 1**: 新增 API 接口文档

  新增以下两份 Markdown 文档：
  - `docs/api-docs/web-server/docs/resource/security_group/batch_overwrite_tcloud_security_group_rule.md`
  - `docs/api-docs/web-server/docs/biz/security-group/batch_overwrite_tcloud_security_group_rule.md`

  格式参考同目录 `batch_update_tcloud_security_group_rule.md`，需注明：接口为全量覆盖语义；
  `egress_rule_set` 和 `ingress_rule_set` 均必填，各自 [1, 100] 条；字段说明与 batch/update 一致。

- [x] **Task 2**: 新增 cloud-server proto 请求结构

  文件：`pkg/api/cloud-server/security_group.go`（或同目录相关文件）

  新增 `TCloudSGRuleOverwriteReq`，`egress_rule_set` 和 `ingress_rule_set` 均必填，validate tag 为
  `required,min=1,max=100`，复用现有 `TCloudSGRuleCreateSpec` 元素类型。

- [x] **Task 3**: 新增 hc-service proto 请求结构

  文件：`pkg/api/hc-service/security_group.go`

  新增 `TCloudSGRuleOverwriteReq`，含 `AccountID`（required）+ `EgressRuleSet`/`IngressRuleSet`
  （各自 `required,min=1,max=100`），复用现有 `TCloudSGRuleCreateSpec`。

- [x] **Task 4**: 新增 TCloud adapter Option 结构

  文件：`pkg/adaptor/types/security-group-rule/tcloud.go`

  新增 `TCloudOverwriteOption`（含 Region、CloudSecurityGroupID、EgressRuleSet、IngressRuleSet），
  实现 `Validate()` 方法。

- [x] **Task 5**: 新增 TCloud adapter 方法

  文件：`pkg/adaptor/tcloud/security_group_rule.go`

  新增 `OverwriteSecurityGroupRule`，调用腾讯云 `ModifySecurityGroupPolicies`，
  `SecurityGroupPolicySet.Version` 不赋值（nil），同时设置 Egress 和 Ingress policies。

- [x] **Task 6**: 新增 hc-service 客户端方法

  文件：`pkg/client/hc-service/tcloud/security_group.go`

  新增 `OverwriteSecurityGroupRule(kt, sgID, req)` 方法，使用 `common.RequestNoResp`，
  路径为 `/security_groups/%s/rules/batch/overwrite`，HTTP method PUT。

- [x] **Task 7**: 新增 hc-service 路由和处理函数

  文件：`cmd/hc-service/service/security-group/tcloud_security_group_rule.go`
  注册到：`cmd/hc-service/service/security-group/security_group.go` 的 `tcloudService`

  新增 `OverwriteTCloudSGRule`：获取 SG 信息 → 构造 TCloudOverwriteOption（Version 不设置）
  → 调用 adapter → 调用 `syncSGRule` 全量同步本地 DB。
  路由：`PUT /vendors/tcloud/security_groups/{security_group_id}/rules/batch/overwrite`

- [x] **Task 8**: 新增 cloud-server 处理函数和路由

  文件：`cmd/cloud-server/service/security-group/overwrite_rule.go`（新文件）
  注册到：`cmd/cloud-server/service/security-group/init.go`

  新增 `OverwriteSecurityGroupRule`（ResOperateAuth）和 `OverwriteBizSGRule`（BizOperateAuth），
  共用内部 `overwriteSGRule`：vendor 校验 → 鉴权 → switch vendor（仅 tcloud） →
  调用 hc-service `OverwriteSecurityGroupRule`。
  路由：
  - 资源下：`PUT /vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`
  - 业务下：`PUT /bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite`
