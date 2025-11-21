<script setup lang="ts">
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { useAttrs } from 'vue';
import DisplayValue from '../display-value/index.vue';

export interface IDataListProps {
  columns?: ModelPropertyColumn[];
  list: any[];
  pagination: PaginationType;
}
defineOptions({ name: 'DataListTable' });
const props = withDefaults(defineProps<IDataListProps>(), {});
const attrs = useAttrs();
const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const { settings } = useTableSettings(props.columns || []);
</script>

<template>
  <bk-table
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :settings="settings"
    remote-pagination
    show-overflow-tooltip
    @page-limit-change="handlePageSizeChange"
    @page-value-change="handlePageChange"
    @column-sort="handleSort"
    v-bind="attrs"
  >
    <slot>
      <bk-table-column
        v-for="(column, index) in columns"
        v-bind="column"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
      >
        <template #default="data">
          <slot :name="column.id" v-bind="data">
            <display-value :property="column" :value="data?.row[column.id]" :display="column?.meta?.display" />
          </slot>
        </template>
      </bk-table-column>
    </slot>
  </bk-table>
</template>
