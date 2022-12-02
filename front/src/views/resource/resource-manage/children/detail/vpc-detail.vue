<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import VPCCidr from '../components/vpc/vpc-cidr.vue';
import VPCRoute from '../components/vpc/vpc-route.vue';
import VPCSubnet from '../components/vpc/vpc-subnet.vue';

import useDeleteVPC from '../../hooks/use-delete-vpc';
import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const {
  isShowVPC,
  handleDeleteVPC,
  DeleteVPC,
} = useDeleteVPC();

const VPCFields = [
  {
    name: 'ID',
    value: '1234223',
  },
  {
    name: '账号ID',
    value: '1234223',
    link: 'http://www.baidu.com',
  },
  {
    name: '账号名称',
    value: '1234223',
  },
  {
    name: '备注',
    value: '1234223',
    edit: true,
  },
  {
    name: 'VPC',
    value: '1234223',
    copy: '1234223',
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
</script>

<template>
  <detail-header>
    VPC：（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleDeleteVPC"
      >
        {{ t('删除') }}
      </bk-button>
    </template>
  </detail-header>

  <detail-info
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

  <DeleteVPC
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
</style>
