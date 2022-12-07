<script setup lang="ts">
import type {
  PlainObject,
} from '@/typings/resource';

import {
  h,
} from 'vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
} from 'vue-router';
import useBusiness from '../../hooks/use-business';

// use hooks
const {
  t,
} = useI18n();

const router = useRouter();

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();

// 状态
const columns = [
  {
    type: 'selection',
  },
  {
    label: 'ID',
    field: '',
    sort: true,
    render({ cell }: PlainObject) {
      return h(
        'span',
        {
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'image',
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
const tableData: any[] = [{}];

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
  </section>

  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
    @column-sort="handleSortBy"
  />

  <resource-business
    v-model:is-show="isShowDistribution"
    :title="t('镜像分配')"
  />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
