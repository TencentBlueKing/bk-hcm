<script setup lang="ts">
import type {
  // PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import {
  Button,
} from 'bkui-vue';

import {
  PropType,
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
import useQueryList from '../../hooks/use-query-list';
import { HostCloudEnum, CloudType } from '@/typings';

// use hook
const {
  t,
} = useI18n();

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

const router = useRouter();

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'cvms');

console.log('datas', datas);

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
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick() {
            jumpTo('resourceDetail', {
              params: { type: 'host' },
              query: {
                id: data.id,
                type: data.vendor,
              },
            });
          },
        },
        [
          data.id || '--',
        ],
      );
    },
  },
  {
    label: '实例 ID',
    field: 'cloud_id',
  },
  {
    label: '云厂商',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          CloudType[data.vendor],
        ],
      );
    },
  },
  {
    label: '地域',
    field: 'region',
  },
  {
    label: '名称',
    field: 'name',
  },
  {
    label: '状态',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          HostCloudEnum[data.status] || data.status,
        ],
      );
    },
  },
  {
    label: '操作系统',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.os_name || '--',
        ],
      );
    },
  },
  {
    label: '云区域ID',
    field: 'bk_cloud_id',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.bk_cloud_id === -1 ? '未分配' : data.bk_cloud_id,
        ],
      );
    },
  },
  {
    label: '内网IP',
    field: '',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.private_ipv4_addresses || data.private_ipv6_addresses,
        ],
      );
    },
  },
  {
    label: '公网IP',
    field: '',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.public_ipv4_addresses || data.public_ipv6_addresses,
        ],
      );
    },
  },
  {
    label: '创建时间',
    field: 'created_at',
  },
  {
    label: '启动时间',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.cloud_launched_time || '--',
        ],
      );
    },
  },
  {
    label: '操作',
    field: '',
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick() {
            jumpTo('resourceDetail', {
              params: { type: 'host' },
              query: {
                id: data.id,
                type: data.vendor,
              },
            });
          },
        },
        [
          '详情',
        ],
      );
    },
  },
];


// 跳转
const jumpTo = (name: string, params?: DoublePlainObject) => {
  router.push({
    name,
    ...params,
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
      :data="datas"
      :pagination="pagination"
      remote-pagination
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
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
  </bk-loading>
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
