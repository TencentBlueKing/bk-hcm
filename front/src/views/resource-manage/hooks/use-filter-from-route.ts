/**
 * 基于 route.query 的 filter hook，用于资源纳管模块
 * - 直接从 route.query 读取 accountId、vendor、assign、searchQs
 * - 一次性构建完整的 filter.rules，不依赖任何 props 传入
 */
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import { buildSearchSelectValueBySearchQsCondition, transformSimpleCondition } from '@/utils/search';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { getSearchProperties } from './search-properties';

function useFilterFromRoute(resourceType: ResourceTypeEnum) {
  const route = useRoute();
  const properties = getSearchProperties(resourceType);

  const searchQs = useSearchQs({
    properties,
    key: 'filter',
  });

  const searchValue = ref<any[]>([]);

  // filter 使用 ref 而非 computed，因为 useQueryList 通过 { filter: filter.value } 持有引用，
  // 只有原地修改 .rules 才能被 useQueryList 的 deep watch 检测到变化。
  const filter = ref<any>({ op: 'and', rules: [] });

  // route.query 变化时，重新计算 filter.rules
  watch(
    () => route.query,
    (query) => {
      const rules: RulesItem[] = [];
      const isImage = resourceType === ResourceTypeEnum.IMAGE;

      // 镜像是公有资源，只需要 vendor + type=public 条件，不需要 accountId / assign
      if (isImage) {
        const vendor = query.vendor as string;
        if (vendor) {
          rules.push({ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor });
        }
        rules.push({ field: 'type', op: QueryRuleOPEnum.EQ, value: 'public' });
      } else {
        // 1. accountId
        const accountId = query.accountId as string;
        if (accountId) {
          rules.push({ field: 'account_id', op: QueryRuleOPEnum.EQ, value: accountId });
        }

        // 2. vendor（仅在 vendor-group 选中厂商、未选具体账号时存在）
        const vendor = query.vendor as string;
        if (vendor) {
          rules.push({ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor });
        }

        // 3. assign（分配状态）
        const assign = query.assign as string;
        if (assign && assign !== 'all') {
          rules.push({
            field: 'bk_biz_id',
            op: Number(assign) === 1 ? QueryRuleOPEnum.NEQ : QueryRuleOPEnum.EQ,
            value: -1,
          });
        }
      }

      // 4. searchQs 搜索条件
      const condition = searchQs.get(query);
      searchValue.value = buildSearchSelectValueBySearchQsCondition(condition, properties);
      const { rules: searchRules = [] } = transformSimpleCondition(condition, properties);
      rules.push(...searchRules);

      // 一次性赋值
      filter.value.rules = rules;
    },
    { immediate: true, deep: true },
  );

  return {
    searchValue,
    filter,
    searchQs,
  };
}

export default useFilterFromRoute;
