<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  watchEffect,
  reactive,
} from 'vue';
import { cloneDeep } from 'lodash';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

let params = reactive(cloneDeep(props));
watchEffect(() => {
  params = cloneDeep(props);
  params.filter.rules = params.filter.rules.filter(e => e.field !== 'account_id');
  params.filter.rules.push({
    field: 'type',
    op: 'eq',
    value: 'public',
  });
});

const columns = useColumns('image');

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(params, 'images');
</script>

<template>
  <bk-loading :loading="isLoading">
    <bk-table
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
</style>
