<script lang="ts" setup>
import {
  ref,
} from 'vue';

const inColumns = [
  {
    label: '来源',
    field: 'id',
  },
  {
    label: '端口协议',
    field: 'id',
  },
  {
    label: '端口',
    field: 'id',
  },
  {
    label: '策略',
    field: 'id',
  },
  {
    label: '操作',
    field: 'id',
  },
];
const inData = [
  {
    id: 233,
  },
];

const outColumns = [
  {
    label: '目标',
    field: 'id',
  },
  {
    label: '端口协议',
    field: 'id',
  },
  {
    label: '端口',
    field: 'id',
  },
  {
    label: '策略',
    field: 'id',
  },
  {
    label: '操作',
    field: 'id',
  },
];
const outData = [
  {
    id: 233,
  },
];
// tab 信息
const types = [
  { name: 'in', label: '入站规则' },
  { name: 'out', label: '出站规则' },
];
const activeType = ref('in');
</script>

<template>
  <section class="mt20 rule-main">
    <bk-radio-group
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

    <bk-button theme="primary">
      新增规则
    </bk-button>
  </section>

  <bk-table
    v-if="activeType === 'in'"
    class="mt20"
    row-hover="auto"
    :columns="inColumns"
    :data="inData"
  />

  <bk-table
    v-if="activeType === 'out'"
    class="mt20"
    row-hover="auto"
    :columns="outColumns"
    :data="outData"
  />
</template>

<style lang="scss" scoped>
  .rule-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
