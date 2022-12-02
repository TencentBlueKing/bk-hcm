<script lang="ts" setup>
import {
  ref,
  h,
} from 'vue';
import {
  Button,
} from 'bkui-vue';

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
    label: '描述',
    field: 'id',
  },
  {
    label: '操作',
    render() {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            console.log(233);
          },
        },
        [
          '解绑',
        ],
      );
    },
  },
];
const tableData = [
  {
    id: 233,
  },
];
const securityGroups = ref([]);
const securityGroup = ref();
</script>

<template>
  <section class="security-head mt20">
    <bk-select
      v-model="securityGroup"
      filterable
    >
      <bk-option
        v-for="item in securityGroups"
        :key="item.value"
        :value="item.value"
        :label="item.label"
      />
    </bk-select>
    <bk-button class="ml10" theme="primary">解绑安全组</bk-button>
  </section>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
</template>

<style lang="scss" scoped>
  .security-head {
    display: flex;
    align-items: center;
  }
</style>
