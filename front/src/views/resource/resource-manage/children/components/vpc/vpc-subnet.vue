<script lang="ts" setup>
import { useRoute } from 'vue-router';
import useColumns from '../../../hooks/use-columns';
import useQueryList from '../../../hooks/use-query-list';

const route = useRoute();
const { columns, settings } = useColumns('subnet');

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange } = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [
        {
          field: 'vpc_id',
          op: 'eq',
          value: route.query.id,
        },
      ],
    },
  },
  'subnets',
);
</script>

<template>
  <bk-loading :loading="isLoading">
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
