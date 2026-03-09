<script setup lang="ts">
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilterFromRoute from '@/views/resource-manage/hooks/use-filter-from-route';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';

const { searchValue, filter, searchQs } = useFilterFromRoute(ResourceTypeEnum.ROUTING);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
  { filter: filter.value },
  'route_tables',
);

const { columns, settings } = useColumns('route');
</script>

<template>
  <bk-loading :loading="isLoading">
    <resource-search-select
      class="search"
      v-model="searchValue"
      :resource-type="ResourceTypeEnum.ROUTING"
      @change="(condition) => searchQs.set(condition)"
    />
    <bk-table
      :settings="settings"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
.mt20 {
  margin-top: 20px;
}
.search {
  margin-left: auto;
}
</style>
