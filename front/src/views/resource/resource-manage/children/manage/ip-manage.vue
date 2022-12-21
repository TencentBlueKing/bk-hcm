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
import useSelection from '../../hooks/use-selection';

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

const {
  selections,
  handleSelectionChange,
} = useSelection();

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
                type: 'ip',
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
const tableData: any[] = [{ id: 1 }, { id: 2 }];

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
    >
      {{ t('释放') }}
    </bk-button>
  </section>

  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
    @column-sort="handleSortBy"
    @selection-change="handleSelectionChange"
  />

  <resource-business
    v-model:is-show="isShowDistribution"
    :data="selections"
    :title="t('弹性IP分配')"
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
