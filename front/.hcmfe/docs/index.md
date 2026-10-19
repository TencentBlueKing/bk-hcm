# hcm 项目文档导航 (Tier A)

> 由 bkdevbuddy docs 维护。动代码前先在这里定位相关模块，再读对应 `modules/<id>.md`。
> 深度随 workflow 逐步补齐（stub → drafted → verified）。

## 基础

- [基础/共享接缝](modules/base.md) — 跨模块共享层：HTTP 客户端、API 封装、全局组件、hooks、store、常量/类型、样式与工具函数等。已拆出 model/menu-route/auth 三个基础能力模块。 · `drafted`  (relates: menu-route, auth)

## 模块

- [组件级多云与变体差异化](modules/component-variant.md) — 单组件同时应对多云与构建变体。典范：account-selector。 · `stub`  (relates: model, plugin-handler, page-variant)
- [页面级变体差异化](modules/page-variant.md) — 用 .plugin 替换整块 UI 片段，承载页面级变体差异。 · `stub`  (relates: component-variant, plugin-handler)
- [菜单与路由](modules/menu-route.md) — 三级一级视图、去中心化路由、菜单与路由解耦、routerAction 统一跳转。 · `drafted`  (relates: base, auth)
- [权限控制](modules/auth.md) — 视图级权限与操作级权限。旧 useVerify/PermissionDialog 已废弃。 · `drafted`  (relates: base, menu-route)
- [操作记录](modules/operation-log.md) — 操作审计日志。 · `stub`
- [资源运营一级视图（容器·待拆解）](modules/resource.md) — 一级视图 /resource/ 容器，聚合账号/回收站/资源纳管，待拆解。 · `drafted`
- [业务一级视图（容器·待拆解）](modules/business.md) — 一级视图 /business/:bizId 容器，待拆解。 · `stub`  (relates: load-balancer)
- [工作台一级视图（容器·待拆解）](modules/service.md) — 一级视图 /service/ 容器，待拆解。 · `stub`
- [云账号管理](modules/cloud-account-manage.md) — 云账号密钥、权限策略/模板、二级/三级账号管理。 · `stub`
- [负载均衡](modules/load-balancer.md) — CLB/监听器/目标组/设备。 · `drafted`  (relates: business)
- [账单](modules/bill.md) — 账单与账号账单视图。 · `drafted`
- [单据/工单](modules/ticket.md) — 工单流程，含业务/服务两套入口。 · `stub`
- [资源选型方案](modules/scheme.md) — 选型方案列表/详情/推荐。 · `stub`
- [任务管理](modules/task.md) — 异步任务（cvm/clb 等）列表与详情。 · `drafted`
- [门户框架/首页/错误页/通知](modules/app-shell.md) — 应用外壳与通用页面：首页、错误页、全局通知。 · `stub`
- [模型驱动与页面模式](modules/model.md) — 装饰器字段元数据模型与多云 Factory、展示/表单/列表三场景模式。 · `drafted`  (relates: menu-route, auth)
- [变体插件模式（退场中）](modules/plugin-handler.md) — 退场中。按变体分目录的历史插件式组织。原则：只迁不增。 · `stub`
