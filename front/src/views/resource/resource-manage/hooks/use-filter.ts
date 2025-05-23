/* eslint-disable no-nested-ternary */
import { onMounted, ref, watch } from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA, SEARCH_VALUE_IDS, VendorEnum } from '@/common/constant';
import cloneDeep from 'lodash/cloneDeep';

import { useAccountStore } from '@/store';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { useRoute } from 'vue-router';
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

const useFilter = (props: PropsType, convertValueCallbacks?: Record<string, (value: any) => any>) => {
  const searchData = ref([]);
  const searchValue = ref([]);
  const filter = ref<any>(cloneDeep(props.filter));
  const isAccurate = ref(false);
  const accountStore = useAccountStore();
  const route = useRoute();
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
              values: [{ id: queryValueItem, name: queryValueItem }],
            })),
          );
        } else {
          params.push({
            id: queryName,
            values: [{ id: queryValue, name: queryValue }],
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
        const convertValueCallback = convertValueCallbacks?.[field];
        if (convertValueCallback) return convertValueCallback(tmpValue);
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

      // 为resource页面下的list页面设置vendor过滤条件（vendor有值的情况：选择了具体的账号或云厂商）
      // 解决的问题：资源下进入下钻页面后，back回来会丢失vendor条件，请求的数据与页面的vendor不一致（比如腾讯云安全组列表页->腾讯云安全组详情页->腾讯云安全组列表页）
      const selectedVendor = resourceAccountStore.vendorInResourcePage;
      // 处理不同场景的过滤规则
      if (selectedVendor) {
        const isGcpSecurity = selectedVendor === VendorEnum.GCP && props.whereAmI === ResourceManageSenario.security;
        if (!isGcpSecurity) {
          const vendorCondition = createVendorCondition(selectedVendor);
          filter.value.rules =
            props.whereAmI === ResourceManageSenario.image
              ? [vendorCondition] // 镜像场景直接设置
              : [...props.filter.rules.filter(({ field }) => field !== 'vendor'), vendorCondition]; // 其他场景合并
        } else {
          // GCP防火墙list页面特殊处理：不添加vendor条件
          filter.value.rules = props.filter.rules.filter(({ field }) => field !== 'vendor');
        }
      } else {
        // 无vendor信息时,除镜像场景，其他均使用原规则
        filter.value.rules = props.whereAmI === ResourceManageSenario.image ? [] : props.filter.rules;
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
