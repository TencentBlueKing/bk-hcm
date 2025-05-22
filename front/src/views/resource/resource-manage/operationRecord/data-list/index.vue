<template>
  <bk-table
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="'calc(100% - 48px)'"
    :settings="settings"
    remote-pagination
    show-overflow-tooltip
    @page-limit-change="handlePageSizeChange"
    @page-value-change="handlePageChange"
    @column-sort="handleSort"
    row-key="id"
  >
    <bk-table-column
      v-for="(column, index) in columns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :sort="column.sort"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
    <bk-table-column :label="'操作'">
      <template #default="{ row }">
        <bk-button theme="primary" text @click="emit('view-detail', row)">查看详情</bk-button>
      </template>
    </bk-table-column>
  </bk-table>
</template>

<script setup lang="ts">
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';

import { ModelProperty, PropertyColumnConfig } from '@/model/typings';
import { IAuditItem } from '@/store/audit';
import { PaginationType } from '@/typings';

const props = defineProps<{
  isBizPage: boolean;
  properties: ModelProperty[];
  list: IAuditItem[];
  pagination: PaginationType;
}>();
const emit = defineEmits<{
  'view-detail': [row: IAuditItem];
}>();

const getColumns = () => {
  const columnIds = ['created_at', 'res_type', 'res_name', 'action', 'source', 'bk_biz_id', 'account_id', 'operator'];
  const columnConfig: Record<string, PropertyColumnConfig> = {
    created_at: {
      sort: true,
    },
    account_id: {
      defaultHidden: true,
    },
  };

  // 业务下不展示业务字段
  const filteredIds = props.isBizPage ? columnIds.filter((id) => id !== 'bk_biz_id') : columnIds;

  return filteredIds.map((id) => ({
    ...props.properties.find((property) => property.id === id),
    ...columnConfig[id],
  }));
};

const columns = getColumns();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const { settings } = useTableSettings(columns);
</script>
