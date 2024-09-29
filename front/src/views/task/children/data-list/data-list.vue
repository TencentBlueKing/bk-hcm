<script setup lang="ts">
import type { IDataListProps } from '@/views/task/typings';
import usePage from '@/hooks/use-page';
import columnFactory from './column-factory';
const { getColumns } = columnFactory();

const props = withDefaults(defineProps<IDataListProps>(), {});

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const columns = getColumns(props.resource);
</script>

<template>
  <div class="task-data-list">
    <bk-table
      ref="tableRef"
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      row-key="id"
    >
      <bk-table-column v-for="(column, index) in columns" :key="index" :prop="column.id" :label="column.name">
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" />
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style lang="scss" scoped>
.task-data-list {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
}
</style>
