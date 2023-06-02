import {
  computed,
  ref,
  watch,
  watchEffect,
} from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA } from '@/common/constant';
import cloneDeep  from 'lodash/cloneDeep';

import {
  useAccountStore,
} from '@/store';
import { QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';
import { useRoute } from 'vue-router';

type PropsType = {
  filter?: FilterType
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
  image = 'image'
};

const useFilter = (props: PropsType) => {
  const searchData = ref([]);
  const searchValue = ref([]);
  const filter = ref<any>(cloneDeep(props.filter));
  const isAccurate = ref(false);
  const accountStore = useAccountStore();
  const route = useRoute();
  const whereAmI = ref<ResourceManageSenario>(route.query['type'] as ResourceManageSenario);

  watchEffect(() => {
    whereAmI.value = route.query['type'] as ResourceManageSenario;
  });

  watch(
    () => whereAmI.value,
    (val) => {
      if(val === ResourceManageSenario.image) filter.value.rules.push(imageInitialCondition);
    }
  );

  watch(
    () => accountStore.accountList,   // 设置云账号筛选所需数据
    (val) => {
      val.length && FILTER_DATA.forEach((e) => {
        if (e.id === 'account_id') {
          e.children = val;
        }
      });
      searchData.value = FILTER_DATA;
      console.log('searchData.value', searchData.value);
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
      if(whereAmI.value === ResourceManageSenario.image) answer.push(imageInitialCondition);
      for (const { id, values } of val) {
        const rule: QueryFilterType = {
          op: QueryRuleOPEnum.OR,
          rules: [],
        };
        const field = id;

        if(whereAmI.value === ResourceManageSenario.image && ['account_id', 'bk_biz_id'].includes(field) ) continue;

        const condition = {
          field,
          op: isAccurate.value ? QueryRuleOPEnum.EQ : QueryRuleOPEnum.CS,
          value:
            field === "bk_cloud_id" ? Number(values[0].id) : values[0].id,
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
      filter.value.rules = answer;
    },
    {
      deep: true,
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