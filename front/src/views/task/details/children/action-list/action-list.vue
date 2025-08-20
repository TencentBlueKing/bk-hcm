<script setup lang="ts">
import { watch, ref } from 'vue';
import get from 'lodash/get';
import { TaskDetailStatus, type IActionListProps } from '@/views/task/typings';
import usePage from '@/hooks/use-page';
import useTableSelection from '@/hooks/use-table-selection';
import { ITaskDetailItem } from '@/store';
import columnFactory from './column-factory';
const props = withDefaults(defineProps<IActionListProps>(), {
  selectable: true,
});

const emit = defineEmits<(e: 'select', data: any[]) => void>();

const { getColumns } = columnFactory();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const isRowSelectEnable = ({ row }: { row: ITaskDetailItem }) => {
  return [TaskDetailStatus.CANCEL, TaskDetailStatus.FAILED].includes(row.state);
};

const { selections, handleSelectAll, handleSelectChange } = useTableSelection({
  isRowSelectable: isRowSelectEnable,
});

// 默认列
const columns = ref(getColumns(props.resource));

watch(
  () => props.detail?.operations,
  (operations) => {
    // 根据操作类型动态生成列
    if (operations) {
      columns.value = getColumns(props.resource, props.detail.operations);
    }
  },
);

watch(
  selections,
  (selections) => {
    emit('select', selections);
  },
  { deep: true },
);
</script>

<template>
  <bk-table
    ref="tableRef"
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="'calc(100vh - 426px)'"
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
    <bk-table-column type="selection" align="center" min-width="30" v-if="selectable"></bk-table-column>
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
          :value="get(row, column.id)"
          :display="column?.meta?.display"
          :vendor="row?.param?.vendor"
        />
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style lang="scss" scoped></style>
