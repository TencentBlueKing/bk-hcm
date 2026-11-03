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

（待补充：由 workflow 在首次改动本模块时深化。）

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
