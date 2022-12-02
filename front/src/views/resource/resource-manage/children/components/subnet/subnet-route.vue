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
  <h3 class="info-title">已绑定路由表：</h3>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
</style>
