# feat-menu-phase3-closing 测试清单

> 工作流: `feat-menu-phase3-closing`
> 关联文档: coding.md（本单无 prd / design / api）

## 验证范围

- 功能边界: `router/module/service.ts` 退场后的路由回归（工作台四类页面仍可进）、`my-apply` / `my-approval` 目录删除后的旧地址与直达地址行为、收藏与分组折叠、终扫命令留证
- 不在范围: 服务请求入口去留（A5，时间触发）、权限改造（权限单）、资源选型/账号管理/单据页面功能本身

## 测试环境

- 前端入口: `对应业务的测试环境地址，由执行人在测试时确认；不要把具体域名写入仓库`
- 测试账号: `<test_account>`（需有单据、权限策略库、资源选型、账号管理的查看权限；另备一个无权限策略库权限的账号）
- 数据准备: 单据列表有历史数据；资源选型开关按需（`ENABLE_CLOUD_SELECTION=true`）

## 用例清单

### P0 - 主流程（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by | 回滚 |
|----|------|------|------|------|------------|------|
| P0-01 | 工作台·单据管理可用 | 已登录 | 工作台 → 个人事务 → 单据管理 | `/service/ticket` 正常打开，列表与页签正常；顶栏高亮工作台 | | 无 |
| P0-02 | 工作台·权限策略库可用 | 有权限账号 | 工作台 → 公共事务 → 权限策略库 | 地址可打开、列表正常；这是原先由 `service.ts` 注册、现在改由 `views/index.ts` 汇总的路由 | | 无 |
| P0-03 | 工作台·资源选型可用 | `ENABLE_CLOUD_SELECTION=true` | 工作台 → 公共事务 → 资源选型 | `/service/scheme/recommendation` 正常，页签与详情跳转正常 | | 无 |
| P0-04 | 工作台·账号管理可用 | 有账号查看权限 | 工作台 → 个人事务 → 账号管理 | `/service/account` 正常，列表无资源账号 | | 无 |
| P0-05 | 旧地址 `/service/my-apply` | 已登录 | 访问 `/service/my-apply`、`/service/my-apply/detail?id=<单据ID>` | 分别转到 `/service/ticket`、`/service/ticket/detail`（对应页面目录已删，redirect 保留） | | 无 |
| P0-06 | 收藏与分组折叠 | 收藏过单据管理（旧收藏含账号管理） | 打开收藏区、折叠再展开分组后刷新 | 无孤儿收藏；分组折叠状态刷新后保持；无重复项 | | 无 |

### P1 - 异常分支（必测）

| ID | 场景 | 前置 | 操作 | 期望 | covered-by |
|----|------|------|------|------|------------|
| P1-01 | 业务侧单据不受影响 | 有业务权限 | 业务资源 → 操作管理 → 单据管理 | 页面与改造前一致，顶栏高亮业务资源 | |
| P1-02 | 权限策略库无权限 | 缺 `AUTH_FIND_PERMISSION_POLICY_LIBRARY` | 展开工作台侧栏 / 直接访问地址 | 菜单隐藏（公共事务若仅剩该项则整组不渲染）；直连显示框架内无权限页 | |
| P1-03 | 已删页面的直达地址 | 已登录 | 访问 `/service/my-approval` | 进 404 状态页（该页面从未注册路由，目录已删，属预期） | |
| P1-04 | 刷新后顶栏顺序不变 | 已登录 | 在任一工作台页面刷新 | 顶栏仍为 业务资源 / 工作台 / 服务请求 / 资源运营（资源选型已不在顶栏） | |

### P2 - 开发侧留证（不需点页面）

- [ ] `rg "MENU_RESOURCE_ACCOUNT_" src` → 0
- [ ] `rg "router/module" src` → 0（`src/router/module/` 目录已删除）
- [ ] `rg "workbenchViews|entry-wb" src` → 0
- [ ] `rg "'scheme-recommendation'|'scheme-list'|'scheme-detail'" src` → 0（CSS 类名除外）
- [ ] `rg "'/scheme" src` → 0
- [ ] `ls src/views/service` → 只剩 `service-apply`

## 验证结论

> 每条用例填写：PASS / FAIL / Skipped + 简要说明（FAIL 必须附复现步骤、错误截图链接或日志片段）

| ID | 结果 | 备注 |
|----|------|------|
| P0-01 | PASS | `/service/ticket` 正常 |
| P0-02 | PASS | 权限策略库正常（路由改由 `views/index.ts` 汇总后无回归） |
| P0-03 | PASS | 资源选型正常，页签与详情跳转正常 |
| P0-04 | PASS | 账号管理正常，列表无资源账号 |
| P0-05 | PASS | `/service/my-apply` 两条旧地址仍转址 |
| P0-06 | PASS | 无孤儿收藏，折叠状态刷新后保持 |
| P1-01 | PASS | 业务侧单据不受影响 |
| P1-02 | PASS | 无权限时菜单隐藏、直连显示框架内无权限页 |
| P1-03 | PASS | `/service/my-approval` 进 404（预期） |
| P1-04 | PASS | 刷新后顶栏顺序不变（资源选型已不在顶栏） |

- 执行日期: 2026-10-27
- 总体结论: PASS
- 后续行动: 文档收口 → 推进 done；`router/module/` 已清空，合入内部版时按 coding.md「合入内部版注意」处理内部版 296 行的 service.ts 与 my-apply 中的 bpass-apply-detail
