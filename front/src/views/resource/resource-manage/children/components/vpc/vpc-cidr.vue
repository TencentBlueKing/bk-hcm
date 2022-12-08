<script lang="ts" setup>
const columns = [
  {
    label: '地址类型',
    field: 'id',
  },
  {
    label: 'CIDR',
    field: 'id',
  },
  {
    label: '状态',
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
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
</template>
