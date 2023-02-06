<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import VPCCidr from '../components/vpc/vpc-cidr.vue';
import VPCRoute from '../components/vpc/vpc-route.vue';
import VPCSubnet from '../components/vpc/vpc-subnet.vue';

import {
  useI18n,
} from 'vue-i18n';
import useColumns from '../../hooks/use-columns';
import useDetail from '../../hooks/use-detail';
import useDelete from '../../hooks/use-delete';

const VPCFields = [
  {
    name: '账号',
    prop: 'account_id',
    link: 'http://www.baidu.com',
  },
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '资源名称',
    prop: 'name',
  },
  {
    name: '业务',
    prop: '1234223',
  },
  {
    name: '云厂商',
    prop: 'vendor',
  },
  {
    name: 'IPv4 CIDR',
    prop: 'ipv4_cidr',
  },
  {
    name: 'IPv6 CIDR',
    prop: 'ipv6_cidr',
  },
  {
    name: '地域 ID',
    prop: 'region',
  },
  {
    name: '地域名称',
    prop: '12312321',
  },
  {
    name: '云区域',
    prop: '1234223',
  },
  {
    name: '状态',
    prop: 'status',
  },
  {
    name: '默认 VPC',
    prop: 'is_default',
  },
  {
    name: '创建时间',
    prop: 'created_at',
  },
  {
    name: '资源组',
    prop: '1234223',
  },
  {
    name: '备注',
    prop: 'memo',
    edit: true,
  },
];
const VPCTabs = [
  {
    name: 'CIDR',
    value: 'cidr',
  },
  {
    name: '子网',
    value: 'subnet',
  },
  {
    name: '路由',
    value: 'route',
  },
];

const {
  t,
} = useI18n();

const columns = useColumns('vpc');

const {
  loading,
  detail,
} = useDetail(
  'vpc',
  '1',
);

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  columns,
  [detail],
  'vpc',
  t('删除 VPC'),
);
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      VPC：（xxx）
      <template #right>
        <bk-button
          class="w100 ml10"
          theme="primary"
          @click="handleShowDelete"
        >
          {{ t('删除') }}
        </bk-button>
      </template>
    </detail-header>

    <detail-info
      :detail="detail"
      :fields="VPCFields"
    />

    <detail-tab
      :tabs="VPCTabs"
    >
      <template #default="type">
        <VPCCidr v-if="type === 'cidr'" />
        <VPCSubnet v-if="type === 'subnet'" />
        <VPCRoute v-if="type === 'route'" />
      </template>
    </detail-tab>

    <delete-dialog>
      {{ t('请注意该VPC包含一个或多个资源，在释放这些资源前，无法删除VPC') }}<br />
      {{ t('子网：{count} 个', { count: 5 }) }}<br />
      {{ t('CVM：{count} 个', { count: 5 }) }}
    </delete-dialog>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
