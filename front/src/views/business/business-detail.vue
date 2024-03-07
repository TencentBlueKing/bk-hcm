<script lang="ts" setup>
import { provide, computed } from 'vue';
import { useRoute } from 'vue-router';

import HostDetail from '@/views/resource/resource-manage/children/detail/host-detail.vue';
import VpcDetail from '@/views/resource/resource-manage/children/detail/vpc-detail.vue';
import SubnetDetail from '@/views/resource/resource-manage/children/detail/subnet-detail.vue';
import SecurityDetail from '@/views/resource/resource-manage/children/detail/security-detail.vue';
import GcpDetail from '@/views/resource/resource-manage/children/detail/gcp-detail.vue';
import DriveDetail from '@/views/resource/resource-manage/children/detail/drive-detail.vue';
import IpDetail from '@/views/resource/resource-manage/children/detail/ip-detail.vue';
import RoutingDetail from '@/views/resource/resource-manage/children/detail/routing-detail.vue';
import ImageDetail from '@/views/resource/resource-manage/children/detail/image-detail.vue';
import NetworkInterfaceDetail from '@/views/resource/resource-manage/children/detail/network-interface-detail.vue';
import TemplateDetail from '../resource/resource-manage/children/detail/template-detail';
import { useVerify } from '@/hooks';
import bus from '@/common/bus';

import { useAccountStore } from '@/store';

const route = useRoute();
const accountStore = useAccountStore();

// 权限hook
const {
  showPermissionDialog,
  handlePermissionConfirm,
  handlePermissionDialog,
  handleAuth,
  permissionParams,
  authVerifyData,
} = useVerify();

const componentMap = {
  host: HostDetail,
  vpc: VpcDetail,
  subnet: SubnetDetail,
  security: SecurityDetail,
  drive: DriveDetail,
  ip: IpDetail,
  routing: RoutingDetail,
  gcp: GcpDetail,
  image: ImageDetail,
  'network-interface': NetworkInterfaceDetail,
  template: TemplateDetail,
};

const renderComponent = computed(() => {
  return Object.keys(componentMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = componentMap[cur];
    return acc;
  }, {});
});

const isResourcePage = computed(() => {
  // 资源下没有业务ID
  return !accountStore.bizs;
});

provide('authVerifyData', authVerifyData); // 将数据传入孙组件
provide('isResourcePage', isResourcePage);

bus.$on('auth', (authActionName: string) => {
  // bus监听
  handleAuth(authActionName);
});
</script>

<template>
  <component :is="renderComponent"></component>
  <permission-dialog
    v-model:is-show="showPermissionDialog"
    :params="permissionParams"
    @cancel="handlePermissionDialog"
    @confirm="handlePermissionConfirm"
  ></permission-dialog>
</template>
