<script lang="ts" setup>
const columns = [
  {
    label: 'ID/名称',
    field: 'id',
  },
  {
    label: 'IP地址',
    field: 'id',
  },
  {
    label: '计费模式',
    field: 'id',
  },
  {
    label: '带宽上限',
    field: 'id',
  },
];
const tableData = [
  {
    id: 233,
  },
];
</script>

<template>
  <bk-table class="mt20" row-hover="auto" :columns="columns" :data="tableData" show-overflow-tooltip />
</template>
