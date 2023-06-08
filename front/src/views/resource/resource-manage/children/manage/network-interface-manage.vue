<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  computed,
} from 'vue';

import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  whereAmI: {
    type: String
  }
});

const columns = useColumns('networkInterface');

const {
  searchData,
  searchValue,
  filter
} = useFilter(props);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList({
  ...props,
  filter: filter.value
}, 'network_interfaces');

const selectSearchData = computed(() => {
  return [
    ...searchData.value,
    ...[{
      name: '公网ipv4',
      id: 'public_ipv4',
    }, {
      name: '内网ipv4',
      id: 'private_ipv4',
    }],
  ];
});


</script>

<template>
  <bk-loading :loading="isLoading">

    <section
      class="flex-row align-items-center mb20 justify-content-end">
      <bk-search-select
        class="w500 ml10"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
      />
    </section>
    <bk-table
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
</style>
