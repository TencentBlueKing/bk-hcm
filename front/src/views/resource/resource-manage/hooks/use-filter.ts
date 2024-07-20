/* eslint-disable no-nested-ternary */
import { onMounted, ref, watch } from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA, SEARCH_VALUE_IDS, VendorEnum } from '@/common/constant';
import cloneDeep from 'lodash/cloneDeep';

import { useAccountStore } from '@/store';
import { QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';
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
    () => searchValue.value,
    (val) => {
      const map = new Map<string, number>();
      const answer = [] as unknown as Array<QueryFilterType | RulesItem>;
      if (props.whereAmI === ResourceManageSenario.image) answer.push(imageInitialCondition);
      for (const { id, values } of val) {
        const rule: QueryFilterType = {
          op: QueryRuleOPEnum.OR,
          rules: [],
        };
        const field = id;

        if (props.whereAmI === ResourceManageSenario.image && ['account_id', 'bk_biz_id'].includes(field)) continue;

        const conditionValue =
          field === 'bk_cloud_id'
            ? Number(values[0].id)
            : field === 'region'
            ? regionStore.getRegionNameEN(values[0].id)
            : values[0].id;
        const condition = {
          field,
          value: conditionValue,
          op:
            field === 'cloud_vpc_ids'
              ? 'json_contains'
              : typeof conditionValue === 'number'
              ? QueryRuleOPEnum.EQ
              : QueryRuleOPEnum.CS,
        };

        if (!map.has(field)) {
          const idx = answer.length;
          map.set(field, idx);
        }
        const idx = map.get(field);
        if (!!answer[idx]) (answer[idx] as QueryFilterType).rules.push(condition);
        else {
          rule.rules.push(condition);
          answer.push(rule);
        }
      }
      if (regionStore.vendor) {
        answer.push({
          field: 'vendor',
          op: QueryRuleOPEnum.EQ,
          value: regionStore.vendor,
        });
      }
      if (props.whereAmI === ResourceManageSenario.image) {
        filter.value.rules = [];
        if (whereAmI.value === Senarios.resource && resourceAccountStore.resourceAccount?.vendor) {
          filter.value.rules = [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: resourceAccountStore.resourceAccount.vendor,
            },
          ];
        }
      } else filter.value.rules = props.filter.rules;
      if (resourceAccountStore.currentVendor === VendorEnum.GCP && props.whereAmI === ResourceManageSenario.security) {
        filter.value.rules = filter.value.rules.filter((e: any) => e.field !== 'vendor');
      }
      filter.value.rules = filter.value.rules.concat(answer);
    },
    {
      deep: true,
      immediate: true,
    },
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
