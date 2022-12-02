<script lang="ts" setup>
import {
  ref,
} from 'vue';
import InfoList from '../../../common/info-list/info-list';

const settingInfo = [
  {
    name: '操作系统',
    value: 'win',
  },
  {
    name: 'CPU',
    value: '100',
  },
  {
    name: '内存',
    value: '100',
  },
  {
    name: '镜像 ID',
    value: '100',
  },
  {
    name: '镜像名称',
    value: 'mirror.tencent',
  },
  {
    name: '镜像类型',
    value: 'image',
  },
];
const columns = [
  {
    label: '目的地址',
    field: 'id',
  },
  {
    label: '下一跳类型',
    field: 'id',
  },
  {
    label: '下一跳',
    field: 'id',
  },
  {
    label: '类型',
    field: 'id',
  },
  {
    label: '备注',
    field: 'id',
  },
];
const tableData = ref([]);
</script>

<template>
  <info-list class="mt20" :fields="settingInfo"></info-list>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
</template>
