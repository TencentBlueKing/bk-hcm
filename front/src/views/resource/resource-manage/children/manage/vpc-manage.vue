<script setup lang="ts">
import type { DoublePlainObject } from '@/typings/resource';

import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';
import useFilterFromRoute from '@/views/resource-manage/hooks/use-filter-from-route';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';

const props = defineProps({
  isResourcePage: {
    type: Boolean,
  },
});

const { selections, handleSelectionChange, resetSelections } = useSelection();

const { columns, settings } = useColumns('vpc');
const { searchValue, filter, searchQs } = useFilterFromRoute(ResourceTypeEnum.VPC);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'vpcs',
);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};

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

const renderColumns = [...columns];

defineExpose({ fetchComponentsData });
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section class="toolbar" :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
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
      <div class="search-selector-container">
        <resource-search-select
          v-model="searchValue"
          :resource-type="ResourceTypeEnum.VPC"
          @change="(condition) => searchQs.set(condition)"
        />
      </div>
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
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-selector-container {
  margin-left: auto;
}
</style>
