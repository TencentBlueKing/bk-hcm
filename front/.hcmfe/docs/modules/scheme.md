# 资源选型方案

> status: stub · kind: module
> globs: `src/views/scheme/**`

选型方案列表/详情/推荐。

## 职责

资源选型（推荐 / 选型方案列表 / 方案详情）与页内方案选择器。三期起本模块是「工作台 → 公共事务」下的一个菜单项，**不再是一级菜单**：路由由 `src/views/scheme/route-config.ts` 导出并并入 `views/index.ts` 的 `serviceViews`，地址前缀 `/service/scheme`，一级菜单 id 复用 `MENU_SCHEME`（`meta.owner = MENU_SERVICE`、`meta.activeKey = MENU_SCHEME`）。

## 关键流程 / 注意事项

- **路由**：父记录 `name: MENU_SCHEME`、`path: 'scheme'`、`redirect: MENU_SCHEME_RECOMMENDATION`；子记录 `recommendation` / `deployment/list` / `deployment/detail`，name 为 `MENU_SCHEME_RECOMMENDATION` / `MENU_SCHEME_LIST` / `MENU_SCHEME_DETAIL`。旧的 `router/module/scheme.ts`、`schemeViews` 已删除，`router/index.ts` 不再有 `MENU_SCHEME` 顶层记录。
- **地址不再兼容旧前缀**：`/scheme/*` 已失效（与 PR #1897 一致，站内 0 处引用）；如需兼容需另加 redirect。
- **页内跳转一律按 name**：`scheme/index.tsx` 页签、`scheme-list`、`scheme-detail`、`components/scheme-selector` 都跳菜单符号，不要写 `/scheme/...` 字符串地址。
- **详情用 query 传参**（`?sid=`），未改路径参数；三页 `breadcrumb.show = false` 沿用现网。
- **权限**：推荐页 `AUTH_CREATE_CLOUD_SELECTION_SCHEME`、列表页 `AUTH_FIND_CLOUD_SELECTION_SCHEME`，菜单显隐仍由 `ENABLE_CLOUD_SELECTION` 控制（`menu-service.ts` 的 `visibility`）。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
