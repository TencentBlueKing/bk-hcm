<script lang="ts" setup>
import {
  ref,
} from 'vue';

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
    label: '可用区',
    field: 'id',
  },
  {
    label: '网段',
    field: 'id',
  },
];
const tableData = ref([]);
</script>

<template>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
</template>
