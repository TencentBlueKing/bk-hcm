<script setup lang="ts">
import type {
  PlainObject,
} from '@/typings/resource';

import {
  h,
} from 'vue';
import {
  useRouter,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useSteps from '../../hooks/use-steps';
import useDeleteVPC from '../../hooks/use-delete-vpc';

// use hooks
const {
  t,
} = useI18n();

const router = useRouter();

const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const {
  isShowVPC,
  handleDeleteVPC,
  DeleteVPC,
} = useDeleteVPC();

// 状态
const columns = [
  {
    type: 'selection',
  },
  {
    label: 'ID',
    field: 'id',
    sort: true,
    render({ cell }: PlainObject) {
      return h(
        'span',
        {
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'vpc',
              },
            });
          },
        },
        [
          cell || '--',
        ],
      );
    },
  },
  {
    label: '实例 ID',
    field: '',
    sort: true,
  },
  {
    label: '名称',
    field: '',
    sort: true,
  },
  {
    label: '云厂商',
    field: '',
    sort: true,
  },
  {
    label: 'IP',
    field: '',
    sort: true,
  },
  {
    label: '云区域',
    field: '',
  },
  {
    label: '地域',
    field: '',
    sort: true,
  },
  {
    label: 'VPC',
    field: '',
    sort: true,
  },
  {
    label: '子网',
    field: '',
    sort: true,
  },
  {
    label: '状态',
    field: '',
  },
  {
    label: '创建时间',
    field: '',
  },
  {
    label: '操作',
    field: '',
  },
];
const tableData: any[] = [{
  id: 233,
}];

// 方法
const handleSortBy = () => {

};
</script>

<template>
  <section>
    <bk-button
      class="w100"
      theme="primary"
      @click="handleDistribution"
    >
      {{ t('分配') }}
    </bk-button>
    <bk-button
      class="w100 ml10"
      theme="primary"
      @click="handleDeleteVPC"
    >
      {{ t('删除') }}
    </bk-button>
  </section>

  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
    @column-sort="handleSortBy"
  />

  <resource-distribution
    v-model:is-show="isShowDistribution"
    :hide-relate-v-p-c="true"
    :title="t('VPC 分配')"
  />

  <delete-VPC
    v-model:is-show="isShowVPC"
  />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
