<script setup lang="ts">
import { useTemplateRef } from 'vue';
import ITree from './tree/index.vue';
import ISearch from './search/index.vue';
import { RulesItem } from '@/typings';

const treeRef = useTemplateRef<typeof ITree>('tree');
const handleSearch = (rules: RulesItem[], searchValue: string) => {
  treeRef.value.search(rules, searchValue);
};
</script>

<template>
  <ISearch @search="handleSearch" />
  <ITree ref="tree" />
</template>

<style scoped></style>
