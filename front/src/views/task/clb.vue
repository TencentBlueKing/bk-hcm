<script setup lang="ts">
import { ref, watch } from 'vue';
import { ResourceTypeEnum } from '@/common/resource-constant';
import useSearchQs from '@/hooks/use-search-qs';
import type { ISearchConditon } from './typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';

const _searchQs = useSearchQs();

const condition = ref<ISearchConditon>({});

const handleSearch = (_vals) => {
  // console.log(vals, 'search');
};

watch(
  condition,
  (_val) => {
    // console.log('watch condition', val);
  },
  { deep: true },
);
</script>

<template>
  <div>
    <search :resource="ResourceTypeEnum.CLB" :condition="condition" @search="handleSearch"></search>
    <data-list></data-list>
  </div>
</template>
