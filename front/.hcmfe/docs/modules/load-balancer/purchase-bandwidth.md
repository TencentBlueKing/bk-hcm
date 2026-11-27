# 购买表单的规格与带宽联动

购买 CLB 表单由 `src/views/service/service-apply/clb/hooks/useRenderForm.tsx` 编排，规格弹框为 `src/views/business/load-balancer/components/lb-spec-type-dialog/index.tsx`。

- 弹框确认返回 `slaType`、`slaName` 和可选的 `bandwidthLimit`。带宽值来自 `src/constants/clb.ts` 的 `CLB_SPEC_TYPE_COLUMNS_MAP`。
- 表单按当前 sla_type 读取上限：共享型及空规格使用 shared.bandwidthLimit（10240）；具体规格优先读取该规格的确认缓存，再回退规格常量和共享型默认值。
- 按规格缓存避免配置清单编辑时复用另一条目的上限；未确认过的既有规格可直接使用常量表。
- Slider 的 maxValue 与 customContent 共用 internetMaxBandwidthOutConfig，范围与刻度同步变化。
- max 变化后，watch 仅收敛超限的 internet_max_bandwidth_out，不主动放大已有带宽。
- 接口提供规格标识与可用性。空库存、不可用规格、共享型、`clb.c1.small` 和 `clb.c5.*` 沿用现有购买限制。只有当前可用列表中的规格可以确认。
- 其他可用但参数表未收录的规格显示接口 `SpecType`，缺失数值显示 `--`。表单在缺失带宽参数时使用 10240 作为兼容默认值。
