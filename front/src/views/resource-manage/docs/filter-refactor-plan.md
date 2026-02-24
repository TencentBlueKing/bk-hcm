# 资源纳管 Filter 重构计划

## 一、迁移可行性结论

### use-filter.ts 引用范围
| 文件 | 模块 |
|------|------|
| security-manage.vue | resource-manage |
| vpc-manage.vue | resource-manage |
| load-balancer-manage.vue | resource-manage |
| subnet-manage.vue | resource-manage |
| drive-manage.vue | resource-manage |
| ip-manage.vue | resource-manage |
| network-interface-manage.vue | resource-manage |
| routing-manage.vue | resource-manage |
| image-manage.vue | resource-manage |

**结论**：useFilter 仅被 resource-manage 下的 manage 组件使用，可迁移。

### use-filter-host.ts 引用范围
| 文件 | 模块 |
|------|------|
| host-manage.vue | resource-manage |
| host-manage.vue | **views/business** |

**结论**：useFilterHost 被 business 模块的 host-manage 使用，迁移后需更新 business 的 import 路径。

### 迁移策略
- 在 `views/resource-manage/hooks/` 下创建**新的** filter hooks（不修改原文件）
- 原 `views/resource/resource-manage/hooks/use-filter.ts` 和 `use-filter-host.ts` 标记为 `@deprecated`
- business 的 host-manage 继续使用原 useFilterHost（或后续迁移时更新 import）

---

## 二、数据流设计

```
route.query (accountId, vendor, filter)
    ↓
resource-search-select (change) → searchQs.set(condition) → route.query.filter
    ↓
新 useFilterFromRoute: watch route.query → 构建 filter (props.filter + searchQs.get 的 rules)
    ↓
useQueryList(filter) → 数据查询
```

---

## 三、searchValue 与 searchQs 格式转换

- **resource-search-select 值**：`ISearchValue[]` = `[{ id, name, values: [{ id, name }] }]`
- **searchQs.set 入参**：`Record<string, string | number | string[] | number[]>`
- **转换**：使用 `getSimpleConditionBySearchSelect`，单值用标量，多值用数组

---

## 四、实施步骤

1. ✅ 扩展 resource-search-select 的 option-common 支持 VPC、SUBNET、DISK、EIP、IMAGE 等
2. ✅ resource-search-select 添加 `@change` 事件，内部使用 route.query 替代 useResourceAccountStore
3. ✅ 创建 `views/resource-manage/hooks/use-filter-from-route.ts` 消费 route.query
4. 🔄 host-manage 已完成迁移；其余 manage 组件待替换
5. ✅ 原 use-filter/use-filter-host 已标记 @deprecated

## 五、已完成改动（host-manage 示例）

- host-manage.vue: useFilterHost → useFilterFromRoute，resource-search-select 添加 @change="(condition) => searchQs.set(condition)"
- 新增 searchSelectValueToSearchQsCondition 工具函数
- 新增 search-properties.ts 定义各资源类型的 ModelPropertyGeneric
