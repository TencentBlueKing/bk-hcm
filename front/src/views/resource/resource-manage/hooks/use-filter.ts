import {
  ref,
  watch,
} from 'vue';

import type { FilterType } from '@/typings/resource';
import { FILTER_DATA } from '@/common/constant';

type PropsType = {
  filter?: FilterType
};


export default (props: PropsType) => {
  const searchData = ref(FILTER_DATA);
  const searchValue = ref([]);
  const filter = ref<any>([]);
  const isAccurate = ref(false);


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
