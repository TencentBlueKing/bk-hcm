<script setup lang="ts">
import type {
  PlainObject,
  DoublePlainObject,
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
import {
  AngleRight,
} from 'bkui-vue/lib/icon';
import useSteps from '../../hooks/use-steps';
import useShutdown from '../../hooks/use-shutdown';
import useReboot from '../../hooks/use-reboot';
import usePassword from '../../hooks/use-password';
import useRefund from '../../hooks/use-refund';
import useBootUp from '../../hooks/use-boot-up';

// use hook
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
  isShowShutdown,
  handleShutdown,
  HostShutdown,
} = useShutdown();

const {
  isShowReboot,
  handleReboot,
  HostReboot,
} = useReboot();

const {
  isShowPassword,
  handlePassword,
  HostPassword,
} = usePassword();

const {
  isShowRefund,
  handleRefund,
  HostRefund,
} = useRefund();

const {
  isShowBootUp,
  handleBootUp,
  HostBootUp,
} = useBootUp();

// 更多
const moreOperations = [
  {
    name: t('重启'),
    handler: handleReboot,
  },
  {
    name: t('重置密码'),
    handler: handlePassword,
  },
  {
    name: t('退回'),
    handler: handleRefund,
  },
];
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
            jumpTo('resourceDetail', { params: { type: 'host' } });
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

// 排序
const handleSortBy = () => {};

// 跳转
const jumpTo = (name: string, params?: DoublePlainObject) => {
  router.push({
    name,
    ...params,
  });
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
      @click="handleBootUp"
    >
      {{ t('开机') }}
    </bk-button>
    <bk-button
      class="w100 ml10"
      theme="primary"
      @click="handleShutdown"
    >
      {{ t('关机') }}
    </bk-button>
    <bk-dropdown
      class="ml10"
      placement="right-start"
    >
      <bk-button>
        <span class="w60">
          {{ t('更多') }}
        </span>
        <angle-right
          width="16"
          height="16"
        />
      </bk-button>
      <template #content>
        <bk-dropdown-menu>
          <bk-dropdown-item
            v-for="operation in moreOperations"
            :key="operation.name"
            @click="operation.handler"
          >
            {{ operation.name }}
          </bk-dropdown-item>
        </bk-dropdown-menu>
      </template>
    </bk-dropdown>
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
    :title="t('主机分配')"
  />

  <host-shutdown
    v-model:isShow="isShowShutdown"
    :title="t('关机')"
  />

  <host-reboot
    v-model:isShow="isShowReboot"
    :title="t('重启')"
  />

  <host-password
    v-model:isShow="isShowPassword"
    :title="t('修改密码')"
  />

  <host-refund
    v-model:isShow="isShowRefund"
    :title="t('主机回收')"
  />

  <host-boot-up
    v-model:isShow="isShowBootUp"
    :title="t('开机')"
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
