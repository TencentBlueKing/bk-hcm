# 工作台一级视图（容器·待拆解）

> status: stub · kind: module
> globs: `src/views/service/**`

一级视图 /service/ 容器，待拆解。

## 职责

工作台 `/service/` 下的个人事务入口。购买 VPC 在服务申请里。

## VPC 购买提单

入口 `src/views/service/service-apply/vpc/index.tsx`：VPC 网段写入 `formData.ipv4_cidr`，初始子网网段写入 `formData.subnet.ipv4_cidr`。修改 VPC 网段时，子网网段会被重置成同一个值，之后用户仍可把子网改得更细。

提交在 `src/views/service/service-apply/hooks/use-vpc-form-data.ts` 的 `getSaveData`。VPC 网段提交 `formData.ipv4_cidr`。腾讯云、华为云、Azure、GCP 的初始子网网段提交 `subnet.ipv4_cidr`。华为云网关地址取子网网段前三段再加 `.1`。AWS 这条提交不带初始子网。后端两个字段都是 CIDR 字符串，并要求子网落在 VPC 网段内。

## 关键流程 / 注意事项

- **视图归属**：本视图即三期起的「工作台」（一级菜单 `MENU_SERVICE`，地址前缀 `/service`）；单据、资源选型、账号管理等页面通过 `views/index.ts` 的 `serviceViews` 汇总，**不新建 `workbenchViews`**，也不在本目录建聚合用的 `route-config.ts`。
- **目录现状**：`src/views/service/` 现在只剩 `service-apply/`（VPC / 子网 / 云主机 / 硬盘 / 负载均衡的提单页）。原 `my-apply/`（13 个文件）与 `my-approval/`（2 个文件）已在三期收尾删除 —— 两者都没有任何路由加载（`/service/my-apply` 只剩 redirect 到 `/service/ticket`），`my-approval` 更从未接过 ITSM 审批接口。注意内部版 `my-apply/components/bpass-apply-detail/` 仍被单据详情引用，合入内部版时不能跟着删。
- **提单成功后的跳转**：云主机 / 硬盘分别由 `hooks/use-vpc-form-data.ts`、`hooks/use-disk-form-data.ts` 处理，统一用 `routerAction.redirect({ name: MENU_SERVICE_TICKET_MANAGEMENT }, { history: true })`；不要写 `/service/my-apply` 这类旧地址（旧地址只剩兼容转址）。资源侧上下文仍走 `router.back()`。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
