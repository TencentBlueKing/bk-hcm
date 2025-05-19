<script setup lang="ts">
import type { DoublePlainObject, FilterType } from '@/typings/resource';

import { PropType, defineExpose, computed } from 'vue';
import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';

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

const { selections, handleSelectionChange, resetSelections } = useSelection();

// use hooks
// const { t } = useI18n();
const { columns, settings } = useColumns('vpc');
const { searchData, searchValue, filter } = useFilter(props);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'vpcs',
);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};
defineExpose({ fetchComponentsData });

const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (!props.isResourcePage) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};

const hostSearchData = computed(() => {
  return [
    {
      name: 'VPC ID',
      id: 'cloud_id',
    },
    ...searchData.value,
    {
      name: '管控区域',
      id: 'bk_cloud_id',
    },
  ];
});

const renderColumns = [...columns];
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.vpcs"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <bk-search-select
        class="w500 ml10 search-selector-container"
        clearable
        :conditions="[]"
        :data="hostSearchData"
        v-model="searchValue"
        value-behavior="need-key"
      />
    </section>

    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
      :is-row-select-enable="isRowSelectEnable"
      show-overflow-tooltip
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.search-selector-container {
  margin-left: auto;
}
</style>
