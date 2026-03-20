<script setup lang="ts">
import { inject, watch } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import { GPU_DEMAND_STATUS, GPU_DEMAND_STATUS_MAP, type IGpuDemandItem } from '@/store/resource-plan/gpu-demand';
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
  adjust: [row: IGpuDemandItem];
  reject: [row: IGpuDemandItem];
  terminate: [row: IGpuDemandItem];
  'start-review': [row: IGpuDemandItem];
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

const canAdjust = (row: IGpuDemandItem) =>
  row.status === GPU_DEMAND_STATUS.INIT ||
  row.status === GPU_DEMAND_STATUS.REJECT ||
  row.status === GPU_DEMAND_STATUS.REJECT_ALL;
const isInitStatus = (row: IGpuDemandItem) => row.status === GPU_DEMAND_STATUS.INIT;
const canReview = (row: IGpuDemandItem) =>
  row.status === GPU_DEMAND_STATUS.PENDING ||
  row.status === GPU_DEMAND_STATUS.REJECT ||
  row.status === GPU_DEMAND_STATUS.REJECT_ALL;
const canReject = (row: IGpuDemandItem) => row.status === GPU_DEMAND_STATUS.PENDING;
const canTerminateSrv = (row: IGpuDemandItem) => row.status === GPU_DEMAND_STATUS.PENDING;
const canTerminateBiz = (row: IGpuDemandItem) =>
  row.status === GPU_DEMAND_STATUS.INIT || row.status === GPU_DEMAND_STATUS.REJECT_ALL;

const getAdjustTip = (row: IGpuDemandItem) => {
  if (canAdjust(row)) return '';
  return `仅待评审、部分已驳回、全部已驳回状态可调整，当前状态：${GPU_DEMAND_STATUS_MAP[row.status]}`;
};

const getTerminateBizTip = (row: IGpuDemandItem) => {
  if (canTerminateBiz(row)) return '';
  return `仅待评审、全部已驳回状态可终止，当前状态：${GPU_DEMAND_STATUS_MAP[row.status]}`;
};

const getTerminateSrvInitTip = () => '待评审状态，请先更新状态至评审中';

const getReviewTip = (row: IGpuDemandItem) => {
  if (canReview(row)) return '';
  if (row.status === GPU_DEMAND_STATUS.DONE) return '已评审';
  if (row.status === GPU_DEMAND_STATUS.TERMINATE) return '已终止';
  return '';
};

const getRejectTip = (row: IGpuDemandItem) => {
  if (canReject(row)) return '';
  if (row.status === GPU_DEMAND_STATUS.TERMINATE) return '已终止';
  return '已评审/已驳回，不支持操作';
};

const getTerminateSrvTip = (row: IGpuDemandItem) => {
  if (canTerminateSrv(row)) return '';
  if (row.status === GPU_DEMAND_STATUS.DONE) return '已评审';
  if (row.status === GPU_DEMAND_STATUS.TERMINATE) return '已终止';
  return '已评审/已驳回，不支持操作';
};
</script>

<template>
  <bk-table
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="`calc(100vh - ${isBusinessPage ? 430 : 516}px)`"
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
        <!-- 需求ID列：蓝色链接样式，点击跳转详情 -->
        <bk-button v-if="column.id === 'id'" theme="primary" text @click="emit('view-details', row)">
          {{ row[column.id] }}
        </bk-button>
        <display-value v-else :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
    <bk-table-column
      :label="'操作'"
      :show-overflow-tooltip="false"
      :min-width="isServicePage ? 150 : 100"
      fixed="right"
    >
      <template #default="{ row }">
        <div class="actions" v-if="isBusinessPage">
          <bk-button
            v-bk-tooltips="{ content: getAdjustTip(row), disabled: canAdjust(row) }"
            theme="primary"
            text
            :disabled="!canAdjust(row)"
            @click="emit('adjust', row)"
          >
            调整
          </bk-button>
          <bk-button
            v-bk-tooltips="{ content: getTerminateBizTip(row), disabled: canTerminateBiz(row) }"
            theme="primary"
            text
            :disabled="!canTerminateBiz(row)"
            @click="emit('terminate', row)"
          >
            终止
          </bk-button>
        </div>
        <div class="actions" v-if="isServicePage && isInitStatus(row)">
          <bk-button theme="primary" text @click="emit('start-review', row)">转为评审中</bk-button>
          <bk-button
            v-bk-tooltips="{ content: getTerminateSrvInitTip(), disabled: false }"
            theme="primary"
            text
            disabled
          >
            终止
          </bk-button>
        </div>
        <div class="actions" v-if="isServicePage && !isInitStatus(row)">
          <bk-button
            v-bk-tooltips="{ content: getReviewTip(row), disabled: canReview(row) }"
            theme="primary"
            text
            :disabled="!canReview(row)"
            @click="emit('view-details', row)"
          >
            去评审
          </bk-button>
          <bk-button
            v-bk-tooltips="{ content: getRejectTip(row), disabled: canReject(row) }"
            theme="primary"
            text
            :disabled="!canReject(row)"
            @click="emit('reject', row)"
          >
            驳回
          </bk-button>
          <bk-button
            v-bk-tooltips="{ content: getTerminateSrvTip(row), disabled: canTerminateSrv(row) }"
            theme="primary"
            text
            :disabled="!canTerminateSrv(row)"
            @click="emit('terminate', row)"
          >
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
