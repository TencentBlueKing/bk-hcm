# Coding — 负载均衡-绑定证书-增加证书ID的展示

## 实施方案

1. 在 `CertSelector` 内以响应式规则实现证书 ID/名称的后端模糊检索（300ms 防抖），保留滚动分页并避免改动共享 `useSingleList`。
2. 组件内缓存已选证书，缺失时通过 `certs/list + cloud_id in` 补查；下拉和暴露给调用方的 `dataList` 均按 `cloud_id` 去重合并当前已选证书与查询结果，保持既有算法校验契约。
3. 下拉以证书 ID 为主，补充名称、算法、过期状态、域名和到期时间；单选/多选回显统一为“证书ID（名称）”。
4. 监听器详情按证书 ID 补查证书信息，逐行展示 ID、名称、算法、到期时间与过期标识，并支持复制 ID。
5. 证书管理仅调整证书专属列、搜索项和删除/单条分配文案，不改共享批量分配组件。

## 影响面分析

- `CertSelector` 全仓仅 `src/views/load-balancer/listener/add.vue` 与 `src/views/load-balancer/listener/layer7/add-domain-dialog.vue` 两个引用文件；保持 props、emits、`dataList`、`handleRefresh` 契约，两个调用方无需改动。
- 不修改共享 `src/hooks/useSingleList.ts`、`src/views/load-balancer/children/display/grid-details.vue`、`src/views/load-balancer/children/display/field-listener.ts`、`src/views/resource/resource-manage/children/dialog/batch-distribution/index.tsx`，避免波及账单、VPC、子网、安全组、CLB、EIP、云硬盘等页面。
- `useColumns('cert')` 仅证书管理页面使用；证书删除/分配仍提交内部 `id`，仅展示 `cloud_id`。
- 证书列表接口现有契约支持嵌套 `or`、`cis` 和 `cloud_id in`，无需后端改动。

## 单据 1

**TAPD**：1069995598138972423

**文件**：
- `src/views/business/load-balancer/clb-view/components/CertSelector/index.tsx`
- `src/views/business/load-balancer/clb-view/components/CertSelector/index.scss`
- `src/views/load-balancer/listener/children/info.vue`
- `src/views/business/cert-manager/index.tsx`
- `src/views/resource/resource-manage/hooks/use-columns.tsx`

**改动点**：证书选择器检索、展示、回显和已选缓存；监听器详情证书信息；证书管理术语与确认文案。

## 验证计划

- lint / prettier / IDE diagnostics。
- E2E mock 覆盖 ID/名称检索、已选补查与回显、过期禁用、详情展示复制、证书管理文案。
- 业务验证结论统一为「待 QA 验证」。
