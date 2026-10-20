# 购买表单的规格与带宽联动

购买 CLB 表单由 `src/views/service/service-apply/clb/hooks/useRenderForm.tsx` 编排，规格弹框为 `src/views/business/load-balancer/components/lb-spec-type-dialog/index.tsx`。

- 弹框确认返回 slaType 与 bandwidthLimit，带宽值来自 `src/constants/clb.ts` 的 CLB_SPEC_TYPE_COLUMNS_MAP。
- 表单按当前 sla_type 读取上限：共享型及空规格使用 shared.bandwidthLimit（10240）；具体规格优先读取该规格的确认缓存，再回退规格常量和共享型默认值。
- 按规格缓存避免配置清单编辑时复用另一条目的上限；未确认过的既有规格可直接使用常量表。
- Slider 的 maxValue 与 customContent 共用 internetMaxBandwidthOutConfig，范围与刻度同步变化。
- max 变化后，watch 仅收敛超限的 internet_max_bandwidth_out，不主动放大已有带宽。
- 本仓库该接口仅提供规格标识与可用性；规格名称优先来自 CLB_SPECS，未命中时弹框与已选预览显示接口 SpecType。
- 参数映射未命中的规格保留在列表，缺失数值显示 --，确认时允许 bandwidthLimit 缺失；表单的 10240 是兼容兜底，不代表未知规格的实际云端上限。
