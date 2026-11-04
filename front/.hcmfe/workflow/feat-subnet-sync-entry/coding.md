# Coding — 子网-增加独立的子网同步入口

## 实施方案

1. 在资源视角与业务视角共用的 `SubnetManage` 工具栏新增“同步子网”按钮，资源视角置于批量分配之后、业务视角自然位于申请按钮之后。
2. 复用 `SyncAccountResource` 同步弹窗：资源视角选定账号时预填账号，未选账号时仍可打开并在弹窗内选择；业务视角传当前业务用于账号过滤；提交资源类型固定为 `subnet`。
3. 资源视角选择具体云厂商/账号时，同步弹窗账号下拉只展示该厂商；选择“全部账号”时展示全部厂商账号。
4. 勾选子网时禁用条件同步入口；提交成功后使用公共弹窗的成功提示，失败时展示后端错误信息或“同步失败”。
5. 为公共 `SyncAccountResource` 增加可选 `vendor` 属性，仅指定时收窄账号列表，不改变现有调用方默认行为；后端 subnet 条件同步能力由 S1/S2 子需求承载。

## 影响面分析

- `SubnetManage` 仅由 `src/views/resource/resource-manage/resource-manage.vue`（资源账号视角）与 `src/views/business/business-manage.vue`（业务视角）使用；两处共用，因此单个入口改动同时覆盖需求截图中的两个位置。
- `src/components/sync-account-resource/index.vue` 新增可选 `vendor` 入参；负载均衡、监听器、安全组等既有调用方不传该属性，账号过滤仍走原默认逻辑，行为不变。
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
