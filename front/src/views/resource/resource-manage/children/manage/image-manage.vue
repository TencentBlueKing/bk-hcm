<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  watch,
  reactive,
  onBeforeUnmount,
  computed,
} from 'vue';
// import { cloneDeep } from 'lodash';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

const params: any = reactive({ filter: { op: 'and', rules: [{
  field: 'type',
  op: 'eq',
  value: 'public',
}] } });
// watchEffect(() => {
//   params = props;
//   params.filter.rules = params.filter.rules.filter(e => e.field !== 'account_id');
//   params.filter.rules.push({
//     field: 'type',
//     op: 'eq',
//     value: 'public',
//   });
// });


const columns = useColumns('image');

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(params, 'images');

const selectSearchData = computed(() => {
  return [
    ...searchData.value,
    ...[{
      name: '公网ipv4',
      id: 'public_ipv4',
    }, {
      name: '内网ipv4',
      id: 'private_ipv4',
    }],
  ];
});

const {
  searchData,
  searchValue,
  isAccurate,
} = useFilter(props);

onBeforeUnmount(() => {
  params.filter.rules = [];
});


// 搜索数据
watch(
  () => searchValue.value,
  (val) => {
    if (val.length) {
      params.filter.rules = val.reduce((p, v) => {
        if (v.type === 'condition') {
          params.filter.op = v.id || 'and';
        } else {
          p.push({
            field: v.id,
            op: isAccurate.value ? 'eq' : 'cs',
            value: v.values[0].id,
          });
        }
        return p;
      }, [
        {
          field: 'type',
          op: 'eq',
          value: 'public',
        },
      ]);
    } else {
      params.filter.rules = [
        {
          field: 'type',
          op: 'eq',
          value: 'public',
        },
      ];
    }
    params.filter.rules = params.filter.rules.filter(e => e.field !== 'account_id' && e.field !== 'bk_biz_id');
  },
  {
    deep: true,
  },
);

// 字段列表
const fieldList: string[] = columns.map(item => item.field);
let dataList: any = datas;
// 接口缺失字段填充默认值
watch(datas, (list) => {
  dataList = list.map(item => {
    fieldList.forEach(field => {
      if (!Object.hasOwnProperty.call(item, field)) {
        item[field] = '--';
      }
    })
    return item;
  })
});
</script>

<template>
  <bk-loading :loading="isLoading">
    <section
      class="flex-row align-items-center mb20 justify-content-end">
      <bk-search-select
        class="w500 ml10"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
      />
    </section>
    <bk-table
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="dataList"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
</style>
