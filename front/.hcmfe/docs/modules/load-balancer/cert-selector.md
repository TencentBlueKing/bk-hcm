# 证书选择与详情展示

证书选择器以证书 ID 为主识别字段，并统一监听器编辑、域名编辑和监听器详情的展示口径。

证书选择器位于 `src/views/business/load-balancer/clb-view/components/CertSelector/index.tsx`，仅由 `src/views/load-balancer/listener/add.vue` 与 `src/views/load-balancer/listener/layer7/add-domain-dialog.vue` 引用。`modelValue` 始终是证书 `cloud_id`，SVR 多选、CA 单选；两个调用方还依赖组件 expose 的 `dataList` 中已选证书的 `encrypt_algorithm` 做算法互斥校验，因此搜索、刷新和分页时不能把已选证书从该列表中丢掉。

组件通过 `certs/list` 做远程检索，基础条件为 `cert_type + account_id`，关键字以嵌套 `or` 同时匹配 `cloud_id cis` 和 `name cis`。已选 ID 不在当前查询结果时，以 `cloud_id in` 补查并缓存；渲染列表按 `cloud_id` 去重合并查询结果与已选证书。下拉展示 ID、名称、算法、过期状态、主/备用域名和到期时间，过期证书仍禁用；回显统一为 `证书ID（名称）`。

监听器基本信息中的证书展示在 `src/views/load-balancer/listener/children/info.vue` 内局部实现：按当前 CA/SVR `cloud_id` 批量补查证书，逐行展示 ID、名称、算法、到期时间与过期标识，并逐项复制 ID。这里不修改共享 `GridDetails` 或监听器字段模型。
