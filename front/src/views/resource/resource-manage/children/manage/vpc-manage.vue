<script setup lang="ts">
import type {
  PlainObject,
  FilterType,
} from '@/typings/resource';

import {
  h,
  PropType,
} from 'vue';
import {
  useRouter,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useSteps from '../../hooks/use-steps';
import useDeleteVPC from '../../hooks/use-delete-vpc';
import useQueryList from '../../hooks/use-query-list';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

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
    label: '资源 ID',
    field: 'cid',
    sort: true,
  },
  {
    label: '名称',
    field: 'name',
    sort: true,
  },
  {
    label: '云厂商',
    field: 'vendor',
  },
  {
    label: '云区域',
    field: 'bk_cloud_id',
  },
  {
    label: '地域',
    field: 'region',
  },
  {
    label: 'IPv4 CIDR',
    field: 'ipv4_cidr',
  },
  {
    label: 'IPv6 CIDR',
    field: 'ipv6_cidr',
  },
  {
    label: '状态',
    field: 'status',
  },
  {
    label: '默认 VPC',
    field: 'is_default',
  },
  {
    label: '子网数',
    field: '',
  },
  {
    label: '创建时间',
    field: 'create_at',
    sort: true,
  },
  {
    label: '操作',
    field: '',
  },
];

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
  handleShowDeleteVPC,
  DeleteVPC,
} = useDeleteVPC();

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'vpc');
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
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
        @click="handleShowDeleteVPC"
      >
        {{ t('删除') }}
      </bk-button>
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>

  <resource-distribution
    v-model:is-show="isShowDistribution"
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
</style>
