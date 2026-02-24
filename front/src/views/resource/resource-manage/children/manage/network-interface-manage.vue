<script setup lang="ts">
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilterFromRoute from '@/views/resource-manage/hooks/use-filter-from-route';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';

const { columns, settings } = useColumns('networkInterface');

const { searchValue, filter, searchQs } = useFilterFromRoute(ResourceTypeEnum.NETWORK_INTERFACE);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
  { filter: filter.value },
  'network_interfaces',
);
</script>

<template>
  <bk-loading :loading="isLoading">
    <resource-search-select
      class="search"
      v-model="searchValue"
      :resource-type="ResourceTypeEnum.NETWORK_INTERFACE"
      @change="(condition) => searchQs.set(condition)"
    />
    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      show-overflow-tooltip
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
.search {
  margin-left: auto;
}
</style>
