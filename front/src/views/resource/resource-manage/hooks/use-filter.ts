/* eslint-disable no-nested-ternary */
import { onMounted, ref, watch } from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA, SEARCH_VALUE_IDS, VendorEnum } from '@/common/constant';
import cloneDeep from 'lodash/cloneDeep';

import { useAccountStore } from '@/store';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { useRoute } from 'vue-router';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useRegionsStore } from '@/store/useRegionsStore';

type PropsType = {
  filter?: FilterType;
  whereAmI?: string;
};

export const imageInitialCondition = {
  field: 'type',
  op: QueryRuleOPEnum.EQ,
  value: 'public',
};

export enum ResourceManageSenario {
  host = 'host',
  vpc = 'vpc',
  subnet = 'subnet',
  security = 'security',
  drive = 'drive',
  interface = 'interface',
  ip = 'ip',
  routing = 'routing',
  image = 'image',
}

const useFilter = (props: PropsType) => {
  const searchData = ref([]);
  const searchValue = ref([]);
  const filter = ref<any>(cloneDeep(props.filter));
  const isAccurate = ref(false);
  const accountStore = useAccountStore();
  const route = useRoute();
  const { whereAmI } = useWhereAmI();
  const resourceAccountStore = useResourceAccountStore();
  const regionStore = useRegionsStore();

  const saveQueryInSearch = () => {
    let params = [] as typeof searchValue.value;
    Object.entries(route.query).forEach(([queryName, queryValue]) => {
      if (!!queryName && SEARCH_VALUE_IDS.includes(queryName)) {
        if (Array.isArray(queryValue)) {
          params = params.concat(
            queryValue.map((queryValueItem) => ({
              id: queryName,
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

  onMounted(() => {
    saveQueryInSearch();
  });

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

  watch(
    () => accountStore.accountList, // 设置云账号筛选所需数据
    (val) => {
      if (!val) return;
      val.length &&
        FILTER_DATA.forEach((e) => {
          if (e.id === 'account_id') {
            e.children = val;
          }
        });
      searchData.value = FILTER_DATA;
    },
    {
      deep: true,
      immediate: true,
    },
  );

  // 搜索数据
  watch(
    searchValue,
    (val) => {
      // 工具函数 - 获取条件值
      const getConditionValue = (field: string, values: any[]): string | number => {
        const tmpValue = values.length > 1 ? values : values[0];
        if (field === 'bk_cloud_id') return Number(values);
        if (field === 'region') return regionStore.getRegionNameEN(tmpValue);
        return tmpValue;
      };

      // 工具函数 - 获取查询操作符
      const getQueryOperator = (field: string, value: unknown): QueryRuleOPEnum => {
        if (field === 'cloud_vpc_ids') return QueryRuleOPEnum.JSON_CONTAINS;
        if (Array.isArray(value)) return QueryRuleOPEnum.IN;
        if (typeof value === 'number' || ['vendor', 'mgmt_type'].includes(field)) return QueryRuleOPEnum.EQ;
        return QueryRuleOPEnum.CS;
      };

      // 工具函数 - 创建厂商条件
      const createVendorCondition = (vendor: string): RulesItem => ({
        field: 'vendor',
        op: QueryRuleOPEnum.EQ,
        value: vendor,
      });

      // 主处理逻辑
      const fieldIndexMap = new Map<string, number>();
      const queryRules: Array<RulesItem> = [];

      // 初始化镜像场景条件
      if (props.whereAmI === ResourceManageSenario.image) {
        queryRules.push(imageInitialCondition);
      }

      // 处理每个搜索项
      val.forEach(({ id: field, values }) => {
        // 跳过禁用字段
        if (props.whereAmI === ResourceManageSenario.image && ['account_id', 'bk_biz_id'].includes(field)) return;

        // 构建条件对象
        const conditionValue = getConditionValue(
          field,
          values.map((e: any) => e.id),
        );
        const condition: RulesItem = {
          field,
          value: conditionValue,
          op: getQueryOperator(field, conditionValue),
        };

        // 分组处理相同字段条件
        if (fieldIndexMap.has(field)) {
          const existingRule = queryRules[fieldIndexMap.get(field)];
          existingRule.rules.push(condition);
        } else {
          const newRule: RulesItem = { op: QueryRuleOPEnum.OR, rules: [condition] };
          fieldIndexMap.set(field, queryRules.length);
          queryRules.push(newRule);
        }
      });

      // 添加云厂商条件
      if (regionStore.vendor) {
        queryRules.push(createVendorCondition(regionStore.vendor));
      }

      // 处理场景规则
      if (props.whereAmI !== ResourceManageSenario.image) {
        filter.value.rules = props.filter.rules;
      } else {
        filter.value.rules = [];
        if (whereAmI.value === Senarios.resource && resourceAccountStore.resourceAccount?.vendor) {
          filter.value.rules = [createVendorCondition(resourceAccountStore.resourceAccount.vendor)];
        }
      }
      if (resourceAccountStore.currentVendor === VendorEnum.GCP && props.whereAmI === ResourceManageSenario.security) {
        filter.value.rules = filter.value.rules.filter((e: RulesItem) => e.field !== 'vendor');
      }

      filter.value.rules = filter.value.rules.concat(queryRules);
    },
    { deep: true, immediate: true },
  );

  watch(
    () => props.filter,
    () => {
      if (/^\/resource\/resource/.test(route.path)) searchValue.value = [];
    },
    {
      deep: true,
      immediate: true,
    },
  );

  return {
    searchData,
    searchValue,
    filter,
    isAccurate,
  };
};

export default useFilter;
