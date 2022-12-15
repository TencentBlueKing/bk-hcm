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
import useDeleteVPC from '../../hooks/use-delete-vpc';
import useDetail from '../../hooks/use-detail';

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
    prop: 'create_at',
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

const {
  isShowVPC,
  handleShowDeleteVPC,
  DeleteVPC,
} = useDeleteVPC();

const {
  loading,
  detail,
} = useDetail(
  'vpc',
  '1',
  VPCFields,
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
          @click="handleShowDeleteVPC"
        >
          {{ t('删除') }}
        </bk-button>
      </template>
    </detail-header>

    <detail-info
      :fields="detail"
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

    <DeleteVPC
      v-model:is-show="isShowVPC"
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
</style>
