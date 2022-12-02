<script lang="ts" setup>
import {
  ref,
} from 'vue';

const cvmColumns = [
  {
    label: '实例 ID',
    field: 'id',
  },
  {
    label: '所属 VPC',
    field: 'id',
  },
  {
    label: '可用区',
    field: 'id',
  },
  {
    label: 'IP 信息',
    field: 'id',
  },
];
const cvmData = [
  {
    id: 233,
  },
];

const networkColumns = [
  {
    label: '实例 ID',
    field: 'id',
  },
  {
    label: '所属 VPC',
    field: 'id',
  },
  {
    label: '可用区',
    field: 'id',
  },
  {
    label: 'IP 信息',
    field: 'id',
  },
  {
    label: '关联主机',
    field: 'id',
  },
];
const networkData = [
  {
    id: 233,
  },
];
// tab 信息
const types = [
  { name: 'cvm', label: 'CVM' },
  { name: 'network', label: '网络接口' },
];
const activeType = ref('cvm');
</script>

<template>
  <bk-radio-group
    class="mt20"
    v-model="activeType"
  >
    <bk-radio-button
      v-for="item in types"
      :key="item.name"
      :label="item.name"
    >
      {{ item.label }}
    </bk-radio-button>
  </bk-radio-group>

  <bk-table
    v-if="activeType === 'cvm'"
    class="mt20"
    row-hover="auto"
    :columns="cvmColumns"
    :data="cvmData"
  />

  <bk-table
    v-if="activeType === 'network'"
    class="mt20"
    row-hover="auto"
    :columns="networkColumns"
    :data="networkData"
  />
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
</style>
