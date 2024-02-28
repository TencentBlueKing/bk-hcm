<script lang="ts" setup>
import useQueryList from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import {
  useRoute,
} from 'vue-router';

const route = useRoute();
const { columns, settings } = useColumns('route');

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
} = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [{
        field: 'vpc_id',
        op: 'eq',
        value: route.query.id,
      }],
    },
  },
  'route_tables',
);
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns.filter((column: any) => !column.onlyShowOnList)"
      :data="datas"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    />
  </bk-loading>
</template>
