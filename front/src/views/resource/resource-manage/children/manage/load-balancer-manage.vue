<template>
  <Loading :loading="isLoading">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.load_balancers"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <Button class="mw88" @click="handleClickBatchDelete" :disabled="selections.length === 0">
        {{ t('批量删除') }}
      </Button>
      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select class="w500" clearable :conditions="[]" :data="clbsSearchData" v-model="searchValue" />
        <slot name="recycleHistory"></slot>
      </div>
    </section>
    <Table
      class="has-selection"
      :columns="columns"
      :data="datas"
      :settings="settings"
      :pagination="pagination"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @column-sort="handleSort"
      row-key="id"
    />
  </Loading>
  <!-- 批量删除负载均衡 -->
  <BatchOperationDialog
    class="batch-delete-lb-dialog"
    v-model:is-show="isBatchDeleteDialogShow"
    :title="t('批量删除负载均衡')"
    theme="danger"
    confirm-text="删除"
    :is-submit-loading="isSubmitLoading"
    :is-submit-disabled="isSubmitDisabled"
    :table-props="tableProps"
    :list="computedListenersList"
    @handle-confirm="handleBatchDeleteSubmit"
  >
    <template #tips>
      已选择
      <span class="blue">{{ tableProps.data.length }}</span>
      个负载均衡，其中
      <span class="red">{{ tableProps.data.filter(({ listenerNum }) => listenerNum > 0).length }}</span>
      个存在监听器不可删除。
    </template>
    <template #tab>
      <BkRadioGroup v-model="radioGroupValue">
        <BkRadioButton :label="true">{{ t('可删除') }}</BkRadioButton>
        <BkRadioButton :label="false">{{ t('不可删除') }}</BkRadioButton>
      </BkRadioGroup>
    </template>
  </BatchOperationDialog>
</template>

<script setup lang="ts">
import { PropType, computed, h } from 'vue';
import { Loading, Table, Button } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { BatchDistribution, DResourceType } from '../dialog/batch-distribution';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import type { DoublePlainObject, FilterType } from '@/typings/resource';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useBatchDeleteLB from '@/views/business/load-balancer/clb-view/all-clbs-manager/useBatchDeleteLB';
import { useI18n } from 'vue-i18n';
import { asyncGetListenerCount } from '@/utils';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  whereAmI: {
    type: String,
  },
});

const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const { searchData, searchValue, filter } = useFilter(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'load_balancers',
  null,
  'list',
  {},
  asyncGetListenerCount,
);
const { selections, handleSelectionChange, resetSelections } = useSelection();
const { columns, settings } = useColumns('lb');

const clbsSearchData = computed(() => [
  {
    name: t('负载均衡ID'),
    id: 'cloud_id',
  },
  ...searchData.value,
]);

const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (whereAmI.value === Senarios.business) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};
// 批量删除负载均衡
const {
  isBatchDeleteDialogShow,
  isSubmitLoading,
  isSubmitDisabled,
  radioGroupValue,
  tableProps,
  handleRemoveSelection,
  handleClickBatchDelete,
  handleBatchDeleteSubmit,
  computedListenersList,
} = useBatchDeleteLB(
  [
    ...columns.slice(1, 7),
    {
      label: '',
      width: 50,
      minWidth: 50,
      render: ({ data }: any) =>
        h(
          Button,
          { text: true, onClick: () => handleRemoveSelection(data.id) },
          h('i', { class: 'hcm-icon bkhcm-icon-minus-circle-shape' }),
        ),
    },
  ],
  selections,
  resetSelections,
  triggerApi,
);
</script>

<style lang="scss" scoped>
.mr15 {
  margin-right: 15px;
}

.search-selector-container {
  margin-left: auto;
}

.batch-delete-lb-dialog {
  :deep(.bkhcm-icon-minus-circle-shape) {
    font-size: 14px;
    color: #c4c6cc;
  }
}
</style>
