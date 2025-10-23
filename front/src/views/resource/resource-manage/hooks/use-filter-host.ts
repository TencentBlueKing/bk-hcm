import { onMounted, ref, watch } from 'vue';
import type { FilterType } from '@/typings/resource';
import { SEARCH_VALUE_IDS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import cloneDeep from 'lodash/cloneDeep';
import { QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';
import { useRoute } from 'vue-router';
import { useRegionsStore } from '@/store/useRegionsStore';
import optionFactory from '@/components/resource-search-select/option-factory';
import { parseIP } from '@/utils';

type PropsType = {
  filter?: FilterType;
  whereAmI?: string;
};

const useFilterHost = (props: PropsType) => {
  const searchValue = ref([]);
  const filter = ref<any>(cloneDeep(props.filter));
  const route = useRoute();
  const regionStore = useRegionsStore();

  const { getOptionData } = optionFactory();
  const filterOptions = getOptionData(ResourceTypeEnum.CVM);

  const saveQueryInSearch = async () => {
    let params = [] as typeof searchValue.value;
    Object.entries(route.query).forEach(([queryName, queryValue]) => {
      if (!!queryName && SEARCH_VALUE_IDS.includes(queryName)) {
        const option = filterOptions.find((item) => item.id === queryName);
        if (Array.isArray(queryValue)) {
          params = params.concat(
            queryValue.map((queryValueItem) => ({
              id: queryName,
              name: option?.name,
              values: [
                {
                  id: queryValueItem,
                  name: queryValueItem,
                },
              ],
            })),
          );
        } else {
          params.push({
            id: queryName,
            name: option?.name,
            values: [
              {
                id: queryValue,
                name: queryValue,
              },
            ],
          });
        }
      }
    });
    searchValue.value = params;
  };

  const getQueryRule = (field: string, rule: Partial<RulesItem>) => {
    const defaultRules: Record<string, Omit<RulesItem, 'field'>> = {
      vendor: {
        op: QueryRuleOPEnum.IN,
        value: [],
      },
      account_id: {
        op: QueryRuleOPEnum.IN,
        value: [],
      },
      cloud_vpc_ids: {
        op: QueryRuleOPEnum.JSON_CONTAINS,
        value: '',
      },
      region: {
        op: QueryRuleOPEnum.CS,
        value: '',
      },
      private_ip: {
        op: QueryRuleOPEnum.JSON_OVERLAPS,
        value: [],
      },
      public_ip: {
        op: QueryRuleOPEnum.JSON_OVERLAPS,
        value: [],
      },
      bk_cloud_id: {
        op: QueryRuleOPEnum.IN,
        value: [],
      },
      bk_asset_id: {
        op: QueryRuleOPEnum.IN,
        value: [],
      },
    };

    const defaultRule = defaultRules[field] ?? {
      op: QueryRuleOPEnum.CS,
      value: '',
    };

    const result = { field, ...defaultRule, ...rule };

    // 按操作符类型格式化值
    const isArrayOp = [QueryRuleOPEnum.IN, QueryRuleOPEnum.JSON_OVERLAPS].includes(result.op);
    if (isArrayOp && !Array.isArray(result.value)) {
      result.value = [result.value] as RulesItem['value'];
    }
    if (!isArrayOp && Array.isArray(result.value)) {
      [result.value] = result.value;
    }

    // 特殊字段值处理
    if (field === 'region') {
      result.value = regionStore.getRegionNameEN(result.value as string);
    }

    // 多IP
    if (field === 'private_ip' || field === 'public_ip') {
      const { IPv4List, IPv6List } = parseIP((result.value as string[]).join(''));
      const IPResult: QueryFilterType = {
        op: QueryRuleOPEnum.OR,
        rules: [],
      };
      if (IPv4List.length) {
        IPResult.rules.push({
          ...result,
          field: field === 'private_ip' ? 'private_ipv4_addresses' : 'public_ipv4_addresses',
          value: IPv4List,
        });
      }
      if (IPv6List.length) {
        IPResult.rules.push({
          ...result,
          field: field === 'private_ip' ? 'private_ipv6_addresses' : 'public_ipv6_addresses',
          value: IPv6List,
        });
      }
      return IPResult;
    }

    return result;
  };

  watch(
    () => route.query,
    () => {
      if (
        Object.entries(route.query)
          .map(([queryName]) => queryName)
          .includes('cloud_id')
      )
        saveQueryInSearch();
    },
    {
      deep: true,
      immediate: true,
    },
  );

  // 搜索数据
  watch(
    () => searchValue.value,
    (val) => {
      let rules = [] as unknown as Array<QueryFilterType | RulesItem>;

      for (const { id: field, values } of val) {
        const ruleValue = values.map((item: any) => item.id);
        const rule = getQueryRule(field, { value: ruleValue });
        rules = rules.concat(rule);
      }

      if (regionStore.vendor) {
        rules.push({
          field: 'vendor',
          op: QueryRuleOPEnum.EQ,
          value: regionStore.vendor,
        });
      }

      // 先用外部传递进来的条件重置
      filter.value.rules = props.filter.rules;

      // 附加新的条件
      filter.value.rules = filter.value.rules.concat(rules);
    },
    {
      deep: true,
      immediate: true,
    },
  );

  // 在资源页面外部条件变化（云厂商或账号）后，需要重新设置查询条件
  watch(
    () => props.filter,
    () => {
      if (/^\/resource\/resource/.test(route.path)) {
        searchValue.value = [];
      }
    },
    {
      deep: true,
      immediate: true,
    },
  );

  onMounted(() => {
    saveQueryInSearch();
  });

  return {
    searchValue,
    filter,
  };
};

export default useFilterHost;
