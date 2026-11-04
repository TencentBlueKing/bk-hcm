# Coding — 子网-增加独立的子网同步入口

## 实施方案

1. 在资源视角与业务视角共用的 `SubnetManage` 工具栏新增“同步子网”按钮，资源视角置于批量分配之后、业务视角自然位于申请按钮之后。
2. 复用 `SyncAccountResource` 同步弹窗：资源视角预填当前账号，业务视角传当前业务用于账号过滤；提交资源类型固定为 `subnet`。
3. 勾选子网时禁用条件同步入口；提交成功后使用公共弹窗的成功提示，失败时展示后端错误信息或“同步失败”。
4. 不修改共享同步组件和后端代码；后端 subnet 条件同步能力由 S1/S2 子需求承载，若环境未上线则在最终报告记录依赖。

## 影响面分析

- `SubnetManage` 仅由 `src/views/resource/resource-manage/resource-manage.vue`（资源账号视角）与 `src/views/business/business-manage.vue`（业务视角）使用；两处共用，因此单个入口改动同时覆盖需求截图中的两个位置。
- `src/components/sync-account-resource/index.vue` 为既有公共组件，本次仅传参，不修改其实现；监听器、安全组等使用方不受影响。
- 子网列表查询、筛选、分页、批量分配逻辑不变。
- 无后端改动。

## 单据 1

**TAPD**：1069995598137973350

**文件**：
- `src/views/resource/resource-manage/children/manage/subnet-manage.vue`

**改动点**：新增资源/业务两视角共用的子网条件同步入口与弹窗接线。

## 验证计划

- lint / IDE diagnostics。
- E2E 覆盖两处入口可见、弹窗打开、勾选态禁用、提交命中 `/resources/subnet/sync_by_cond` 与失败提示。
- 真实环境同步结果待 QA 验证；记录后端 subnet 注册依赖。
