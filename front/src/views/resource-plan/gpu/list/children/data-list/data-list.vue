<script setup lang="ts">
import { inject, watch } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import { GPU_DEMAND_STATUS, type IGpuDemandItem } from '@/store/resource-plan/gpu-demand';
import usePage from '@/hooks/use-page';
import useTableSelection from '@/hooks/use-table-selection';
import useTableSettings from '@/hooks/use-table-settings';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: IGpuDemandItem[];
  pagination: PaginationType;
}

const props = withDefaults(defineProps<IDataListProps>(), {});
const emit = defineEmits<{
  'view-details': [row: IGpuDemandItem];
  reject: [row: IGpuDemandItem];
  terminate: [row: IGpuDemandItem];
  select: [selections: IGpuDemandItem[]];
}>();

const isBusinessPage = inject('isBusinessPage', false);
const isServicePage = inject('isServicePage', false);

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

const isRowSelectEnable = ({ row }: { row: IGpuDemandItem }) => {
  return row.status === GPU_DEMAND_STATUS.INIT;
};

const { selections, handleSelectAll, handleSelectChange } = useTableSelection({
  isRowSelectable: isRowSelectEnable,
});

watch(selections, (val) => emit('select', val), { deep: true });

const canReview = (row: IGpuDemandItem) => row.status !== GPU_DEMAND_STATUS.INIT;
const canReject = (row: IGpuDemandItem) => row.status === GPU_DEMAND_STATUS.PENDING;
const canTerminateSrv = (row: IGpuDemandItem) => row.status === GPU_DEMAND_STATUS.PENDING;
const canTerminateBiz = (row: IGpuDemandItem) =>
  row.status === GPU_DEMAND_STATUS.INIT || row.status === GPU_DEMAND_STATUS.REJECT_ALL;
</script>

<template>
  <bk-table
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="`calc(100vh - ${isBusinessPage ? 500 : 452}px)`"
    :settings="settings"
    :is-row-select-enable="isRowSelectEnable"
    remote-pagination
    show-overflow-tooltip
    @page-limit-change="handlePageSizeChange"
    @page-value-change="handlePageChange"
    @column-sort="handleSort"
    @select-all="handleSelectAll"
    @selection-change="handleSelectChange"
    row-key="id"
  >
    <bk-table-column v-if="isServicePage" type="selection" min-width="30" fixed="left" />
    <bk-table-column
      v-for="(column, index) in columns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :sort="column.sort"
      :min-width="column.minWidth"
      :fixed="column.fixed"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
    <bk-table-column :label="'操作'" :min-width="isServicePage ? 150 : 100">
      <template #default="{ row }">
        <div class="actions" v-if="isBusinessPage">
          <bk-button theme="primary" text @click="emit('view-details', row)">调整</bk-button>
          <bk-button theme="primary" text :disabled="!canTerminateBiz(row)" @click="emit('terminate', row)">
            终止
          </bk-button>
        </div>
        <div class="actions" v-if="isServicePage">
          <bk-button theme="primary" text :disabled="!canReview(row)" @click="emit('view-details', row)">
            评审
          </bk-button>
          <bk-button theme="primary" text :disabled="!canReject(row)" @click="emit('reject', row)">驳回</bk-button>
          <bk-button theme="primary" text :disabled="!canTerminateSrv(row)" @click="emit('terminate', row)">
            终止
          </bk-button>
        </div>
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style lang="scss" scoped>
.actions {
  display: flex;
  gap: 12px;
}
</style>
