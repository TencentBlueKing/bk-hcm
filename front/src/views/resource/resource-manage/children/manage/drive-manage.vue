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
import useMountedDrive from '../../hooks/use-mounted-drive';
import useUninstallDrive from '../../hooks/use-uninstall-drive';

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
  isShowMountedDrive,
  handleMountedDrive,
  MountedDrive,
} = useMountedDrive();

const {
  isShowUninstallDrive,
  handleUninstallDrive,
  UninstallDrive,
} = useUninstallDrive();

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
              params: { type: 'drive' },
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
    <bk-button
      class="w100 ml10"
      theme="primary"
      @click="handleMountedDrive"
    >
      {{ t('挂载') }}
    </bk-button>
    <bk-button
      class="w100 ml10"
      theme="primary"
      @click="handleUninstallDrive"
    >
      {{ t('卸载') }}
    </bk-button>
    <bk-button
      class="w100 ml10"
      theme="primary"
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

  <resource-business
    v-model:is-show="isShowDistribution"
    :title="t('云硬盘分配')"
  />

  <mounted-drive
    v-model:is-show="isShowMountedDrive"
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
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
