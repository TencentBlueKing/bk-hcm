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
  InfoBox,
  Message,
} from 'bkui-vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
} from 'vue-router';
import {
  useResourceStore,
} from '@/store/resource';

import useSteps from '../../hooks/use-steps';
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
                type: 'subnet',
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
    label: '所属 VPC',
    field: 'vpc_cid',
  },
  {
    label: '可用区',
    field: 'zone',
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
    label: '关联路由表',
    field: '',
  },
  {
    label: '状态',
    field: 'status',
  },
  {
    label: '默认子网',
    field: 'is_default',
  },
  {
    label: '可用 IPv4 地址',
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

const resourceStore = useResourceStore();

const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'subnet');

const handleDeleteSubnet = () => {
  InfoBox({
    title: '确认要删除？',
    theme: 'danger',
    onConfirm() {
      return resourceStore
        .delete('subnet', '123')
        .then(() => {
          Message({
            theme: 'success',
            message: '删除成功',
          });
        });
    },
  });
};
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
        @click="handleDeleteSubnet"
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
    :title="t('子网分配')"
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
