<script setup lang="ts">
import { computed, ref, useSlots, watch, watchEffect } from 'vue';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useTableSelection from '@/hooks/use-table-selection';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { SecurityGroupRelResourceByBizItem, SecurityGroupRelatedResourceName } from '@/store/security-group';
import columnFactory from './column-factory';
import { PaginationType } from '@/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';

const { getColumns } = columnFactory();

const props = withDefaults(
  defineProps<{
    resourceName: SecurityGroupRelatedResourceName;
    operation: string;
    list: SecurityGroupRelResourceByBizItem[];
    pagination: PaginationType;
    enableQuery?: boolean;
    hasSelections?: boolean;
    isRowSelectEnable?: (args: { row: any }) => boolean;
    hasSettings?: boolean;
    maxHeight?: string;
  }>(),
  {
    enableQuery: false,
    hasSelections: true,
    hasSettings: true,
  },
);
const emit = defineEmits<(e: 'select', data: any[]) => void>();
const slots = useSlots();
const { whereAmI } = useWhereAmI();

const columns = ref(getColumns(props.resourceName, props.operation));
const { settings } = useTableSettings(columns.value);
const { handlePageChange, handlePageSizeChange, handleSort } = usePage(props.enableQuery, props.pagination);
const { selections, checked, handleSelectAll, handleSelectChange, resetSelections } = useTableSelection({
  rowKey: 'cloud_id',
  isRowSelectable: props.isRowSelectEnable,
});

const tableRef = ref();
const handleClear = () => {
  resetSelections();
  tableRef.value.clearSelection();
};
const handleDelete = (cloud_id: string) => {
  // 移除内存中的勾选项
  const idx = selections.value.findIndex((item) => item.cloud_id === cloud_id);
  selections.value.splice(idx, 1);

  // 更新组件的勾选态
  const tableSelection = tableRef.value.getSelection();
  const row = tableSelection.find((item: SecurityGroupRelResourceByBizItem) => item.cloud_id === cloud_id);
  // TODO：这里手动控制组件行为有点问题，没办法保证 toggleRowSelection 执行后三种状态的视觉效果
  tableRef.value.toggleRowSelection(row, false);
};

const resourceType = computed(() => {
  return ResourceTypeEnum[props.resourceName];
});
watchEffect(() => {
  // 根据操作类型动态生成列
  if (props.resourceName && props.operation) {
    const cols = getColumns(props.resourceName, props.operation);
    if (whereAmI.value === Senarios.business) {
      columns.value = cols.filter((col) => col.id !== 'bk_biz_id');
    } else {
      columns.value = cols;
    }
  }
});

watch(
  selections,
  (val) => {
    if (val.length === 0) tableRef.value.clearSelection(); // 保证通过ref删除时，全不选的视觉效果正常
    emit('select', val);
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
    row-key="cloud_id"
    selection-key="cloud_id"
    :checked="checked"
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
          :resource-type="resourceType"
        />
      </template>
    </bk-table-column>
    <!-- 操作列 -->
    <bk-table-column v-if="slots.operate" :label="'操作'">
      <template #default="{ row }">
        <slot name="operate" :row="row"></slot>
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style lang="scss" scoped></style>
