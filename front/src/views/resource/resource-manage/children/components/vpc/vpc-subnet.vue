<script lang="ts" setup>
import {
  h,
} from 'vue';
import {
  Button,
} from 'bkui-vue';

import {
  useI18n,
} from 'vue-i18n';
import useSteps from '../../../hooks/use-steps';

const {
  t,
} = useI18n();

const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const columns = [
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '资源 ID',
    render() {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            handleDistribution();
          },
        },
        [
          'www',
        ],
      );
    },
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
    label: '默认专有网络',
    field: 'id',
  },
  {
    label: '网段',
    field: 'id',
  },
  {
    label: '关联路由表',
    field: 'id',
  },
  {
    label: '创建时间',
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

  <resource-distribution
    v-model:is-show="isShowDistribution"
    :title="t('子网分配')"
  />
</template>
