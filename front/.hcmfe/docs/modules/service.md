# 工作台一级视图（容器·待拆解）

> status: drafted · kind: module
> globs: `src/views/service/**`

一级视图 /service/ 容器，待拆解。

## 职责

本次已核对的 CLB 购买表单入口为 `src/views/service/service-apply/clb/hooks/useRenderForm.tsx`，编排规格选择与带宽输入；规格弹框位于 business 容器。CLB 业务规则统一记在 [负载均衡购买带宽](load-balancer/purchase-bandwidth.md)，不在容器文档重复维护。其它服务申请职责待后续按需核实。

## 关键流程 / 注意事项

购买表单的带宽范围与刻度共用计算配置，编辑配置清单时按条目的当前规格派生上限；详细优先级及收敛行为见上述负载均衡文档。

带宽收敛监听的是计算上限 `internetMaxBandwidthOutConfig.value.max`，回调参数明确为 number；只在已填值超过新上限时写回表单，配置条目的其它字段保持原有编辑流程。

规格确认载荷允许缺失带宽参数；已选规格名称在静态表未命中时显示接口规格标识。未知规格的处理边界见负载均衡购买带宽文档。

## 明细目录

<!-- bkdevbuddy:toc:start -->
<!-- bkdevbuddy:toc:end -->

（上方标记块由 `bkdevbuddy docs deepen` 按 `modules/<id>/` 目录自动生成，**请勿手工编辑**；拆出 topic 后 deepen 一次即出现链接。）

## 明细约定

优先在本文件用章节深化。**本轮只碰模块的一块、且本文件已有实质章节**时，优先拆 `modules/<id>/<topic>.md`（并行改同一模块时能少撞正文）；某块稳定超 80–120 行是兜底阈值。topic 文件首行写 `# 标题`、紧随一行摘要（TOC 据此渲染）。拆出后勿复制正文。
