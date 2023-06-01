import {
  ref,
  watch,
} from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA } from '@/common/constant';

import {
  useAccountStore,
} from '@/store';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';

type PropsType = {
  filter?: FilterType
};

const accountStore = useAccountStore();


export default (props: PropsType) => {
  const searchData = ref([]);
  const searchValue = ref([]);
  const filter = ref<any>([]);
  const isAccurate = ref(false);

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

  watch(
    () => props.filter,
    (val) => {
      filter.value = val;
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
      const answer = [] as unknown as Array<QueryFilterType>;
      for (const { id, values } of val) {
        const rule: QueryFilterType = {
          op: QueryRuleOPEnum.OR,
          rules: [],
        };
        const field = id;

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
        if (!!answer[idx]) answer[idx].rules.push(condition);
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
