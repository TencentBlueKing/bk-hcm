<script setup lang="ts">
import type { FilterType } from '@/typings/resource';
import { computed, PropType } from 'vue';
// import {
//   useI18n,
// } from 'vue-i18n';

import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  whereAmI: {
    type: String,
  },
});

const { searchData, searchValue, filter } = useFilter(props);

// use hooks
// const {
//   t,
// } = useI18n();
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
  { filter: filter.value },
  'route_tables',
);

const selectSearchData = computed(() => {
  return [
    {
      name: '路由表ID',
      id: 'cloud_id',
    },
    ...searchData.value,
  ];
});

const { columns, settings } = useColumns('route');
</script>

<template>
  <bk-loading :loading="isLoading">
    <section class="flex-row align-items-center justify-content-end">
      <bk-search-select
        class="w500 ml10"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
        value-behavior="need-key"
      />
    </section>
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
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
