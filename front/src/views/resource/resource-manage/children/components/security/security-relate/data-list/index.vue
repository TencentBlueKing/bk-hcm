<script setup lang="ts">
import { ref, watch, watchEffect } from 'vue';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useTableSelection from '@/hooks/use-table-selection';
import { SecurityGroupRelResourceByBizItem, SecurityGroupRelatedResourceName } from '@/store/security-group';
import columnFactory from './column-factory';
import { PaginationType } from '@/typings';

const { getColumns } = columnFactory();

const props = withDefaults(
  defineProps<{
    resourceName: SecurityGroupRelatedResourceName;
    operation: string;
    list: SecurityGroupRelResourceByBizItem[];
    pagination: PaginationType;
    hasSelections?: boolean;
    isRowSelectEnable?: (args: { row: any }) => boolean;
    hasSettings?: boolean;
    maxHeight?: string;
  }>(),
  {
    hasSelections: true,
    hasSettings: true,
  },
);
const emit = defineEmits<(e: 'select', data: any[]) => void>();

const columns = ref(getColumns(props.resourceName, props.operation));
const { settings } = useTableSettings(columns.value);
const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const { selections, handleSelectAll, handleSelectChange, resetSelections } = useTableSelection({
  isRowSelectable: props.isRowSelectEnable,
});

const tableRef = ref();
const handleClear = () => {
  resetSelections();
  tableRef.value.clearSelection();
};
const handleDelete = (cloud_id: string) => {
  // getSelection()获取到的勾选项，不计顺序；selections中是计入勾选顺序的。
  // 通过方法删除勾选项，需要判断删除的是不是最后一项，如果是的话，需要清空一下表格勾选项，保证视觉效果正确。
  const tableSelection = tableRef.value.getSelection();
  const row = tableSelection.find((item: SecurityGroupRelResourceByBizItem) => {
    if (item.cloud_id === cloud_id) {
      const idx = selections.value.findIndex((item) => item.cloud_id === cloud_id);
      selections.value.splice(idx, 1);
      return true;
    }
    return false;
  });
  if (tableSelection.length > 1) {
    tableRef.value.toggleRowSelection(row, false);
  } else {
    handleClear();
  }
};

watchEffect(() => {
  // 根据操作类型动态生成列
  if (props.resourceName && props.operation) {
    columns.value = getColumns(props.resourceName, props.operation);
  }
});

watch(
  selections,
  (selections) => {
    emit('select', selections);
  },
  { deep: true },
);

defineExpose({ handleClear, handleDelete });
</script>

<template>
  <bk-table
    ref="tableRef"
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="maxHeight"
    :settings="hasSettings ? settings : undefined"
    remote-pagination
    show-overflow-tooltip
    :is-row-select-enable="isRowSelectEnable"
    @page-limit-change="handlePageSizeChange"
    @page-value-change="handlePageChange"
    @column-sort="handleSort"
    @select-all="handleSelectAll"
    @selection-change="handleSelectChange"
    row-key="id"
  >
    <!-- 复选框列 -->
    <bk-table-column v-if="hasSelections" width="30" min-width="30" type="selection"></bk-table-column>

    <bk-table-column
      v-for="(column, index) in columns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :sort="column.sort"
      :render="column.render"
    >
      <template #default="{ row }">
        <display-value
          :property="column"
          :value="row[column.id]"
          :display="column?.meta?.display"
          :vendor="row.vendor"
        />
      </template>
    </bk-table-column>
    <!-- 操作列 -->
    <slot name="operate"></slot>
  </bk-table>
</template>

<style lang="scss" scoped></style>
