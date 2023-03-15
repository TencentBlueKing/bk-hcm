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

// use hooks
const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'eips');

const columns = useColumns('eips');
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <slot></slot>

    <bk-table
      class="mt20"
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
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
