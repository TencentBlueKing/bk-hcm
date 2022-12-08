<script lang="ts" setup>
const columns = [
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '名称',
    field: 'id',
  },
  {
    label: '类型',
    field: 'id',
  },
  {
    label: '关联子网数',
    field: 'id',
  },
  {
    label: '计费模式',
    field: 'id',
  },
  {
    label: '创建时间',
    field: 'id',
  },
  {
    label: '过期时间',
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
