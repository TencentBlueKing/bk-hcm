# feat-menu-phase3-scheme 测试清单

> 工作流: `feat-menu-phase3-scheme`
> 关联文档: coding.md（本单无 prd / design / api）

## 验证范围

- 功能边界: 资源选型从顶栏一级菜单移入「工作台 → 公共事务」、路由迁到 `views/scheme/route-config.ts` 并挂 `serviceViews`、地址改为 `/service/scheme/*`、路由 name 符号化、页内跳转跟随
- 不在范围: 资源选型的页面功能与接口改动、部署方案字段调整、账号管理（A4）、机房裁撤（B 线）
- 已知行为变化: `/scheme/*` 旧地址**不再兼容**（与 PR #1897 一致，站内无引用）

## 测试环境

- 前端入口: `对应业务的测试环境地址，由执行人在测试时确认；不要把具体域名写入仓库`
- 测试账号: `<test_account>`（需具备资源选型推荐与方案列表权限；另备一个缺 `FIND_CLOUD_SELECTION_SCHEME` 的账号）
- 数据准备:
  - 方案列表至少有 1 条部署方案（验证明细与返回）
  - 验证开关关闭时需在构建配置里设置 `ENABLE_CLOUD_SELECTION=false`

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 顶栏与侧栏位置 | `ENABLE_CLOUD_SELECTION=true` | 查看顶栏，再展开工作台侧栏 | 顶栏无「资源选型」一级菜单；工作台「公共事务」下有「资源选型」 | | 无 |
| P0-02 | 进入资源选型推荐页 | 已登录 | 工作台 → 公共事务 → 资源选型 | 地址为 `/service/scheme/recommendation`；顶栏高亮工作台；侧栏高亮资源选型 | | 无 |
| P0-03 | 页签切换与刷新 | 在推荐页 | 点「选型方案」页签后刷新 | 地址为 `/service/scheme/deployment/list`；刷新后仍停在选型方案页签 | | 无 |
| P0-04 | 方案详情与返回 | 列表有数据 | 点击任一方案进入详情，再返回列表 | 地址为 `/service/scheme/deployment/detail?sid=…`；返回回到 `/service/scheme/deployment/list`；侧栏高亮资源选型 | | 无 |
| P0-05 | 推荐页创建/保存后跳转 | 有创建权限 | 推荐页完成推荐并保存方案 | 保存成功后跳到选型方案列表（`/service/scheme/deployment/list`），无 404 | | 删除该方案 |
| P0-06 | 方案选择器跳转 | 打开含方案选择器的页面 | 在弹出列表里点某方案 / 点返回 | 分别跳到详情 `sid=…` 与选型方案列表，地址前缀正确 | | 无 |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | 开关关闭时菜单隐藏 | `ENABLE_CLOUD_SELECTION=false` | 查看工作台侧栏 | 「公共事务」下无「资源选型」；若该组因此无可见项，整组不渲染 | |
| P1-02 | 无列表查看权限 | 缺 `AUTH_FIND_CLOUD_SELECTION_SCHEME` | 访问 `/service/scheme/deployment/list` | 框架内显示无权限页，不请求列表数据 | |
| P1-03 | 无推荐权限 | 缺 `AUTH_CREATE_CLOUD_SELECTION_SCHEME` | 访问 `/service/scheme/recommendation` | 显示框架内无权限页 | |
| P1-04 | 旧地址不再兼容 | 已登录 | 地址栏访问 `/scheme/recommendation` | 进 404 状态页（与 PR 一致，本次不做转址）；若产品要求兼容则记 FAIL 并回到设计确认 | |
| P1-05 | 内嵌资源选择器 | 打开含方案的业务页面 | 使用页内资源选型弹层 | 跳转与列表行为正常，地址前缀为 `/service/scheme/*` | |

### P2 - UI 细节（按需）

- [ ] 资源选型三个页面均不显示面包屑（沿用现网 `breadcrumb.show=false`）
- [ ] 详情页以 `sid` 传参的现状未变（未改成路径参数）
- [ ] 收藏「资源选型」后可正常从收藏区进入
- [ ] English 下「资源选型」「选型方案」文案正常

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 | PASS | 顶栏无资源选型，工作台·公共事务下有 |
| P0-02 | PASS | `/service/scheme/recommendation`，顶栏与侧栏高亮正确 |
| P0-03 | PASS | 页签切到 `/service/scheme/deployment/list`，刷新保持 |
| P0-04 | PASS | 详情 `sid` 传参正常，返回回到列表 |
| P0-05 | PASS | 保存方案后跳选型方案列表，无 404 |
| P0-06 | PASS | 选择器跳详情 / 返回列表地址前缀正确 |
| P1-01 | PASS | 开关关闭时组内无资源选型，空组不渲染 |
| P1-02 | PASS | 缺列表权限显示框架内无权限页 |
| P1-03 | PASS | 缺推荐权限显示框架内无权限页 |
| P1-04 | PASS | `/scheme/recommendation` 进 404，与 PR 一致（本期不做转址） |
| P1-05 | PASS | 页内资源选型弹层跳转正常 |

- 执行日期: 2026-10-26
- 总体结论: PASS
- 后续行动: 文档收口 → 推进 done；`/scheme/*` 旧地址按 PR 不兼容，如需兼容另开单补转址
