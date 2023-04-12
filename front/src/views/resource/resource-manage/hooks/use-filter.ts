import {
  ref,
  watch,
} from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA } from '@/common/constant';

import {
  useAccountStore,
} from '@/store';

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
      filter.value.rules = val.reduce((p, v) => {
        if (v.type === 'condition') {
          filter.value.op = v.id || 'and';
        } else {
          p.push({
            field: v.id,
            op: isAccurate.value ? 'eq' : 'cs',
            value: v.values[0].id,
          });
        }
        return p;
      }, []);
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
