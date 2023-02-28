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

const selected = ref('出站规则');
const tableData = [
  {
    id: 233,
  },
];
const handleSelected = (v: string) => {
  selected.value = v;
};
</script>

<template>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />
  <bk-button-group class="mt20">
    <bk-button
      :selected="selected === '出站规则'"
      @click="handleSelected('出站规则')"
    >
      出站规则
    </bk-button>
    <bk-button
      :selected="selected === '入站规则'"
      @click="handleSelected('入站规则')"
    >
      入站规则
    </bk-button>
  </bk-button-group>
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
