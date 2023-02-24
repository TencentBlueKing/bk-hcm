<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
} from 'vue';

import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

const columns = useColumns('networkInterface');

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'network_interfaces');
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
