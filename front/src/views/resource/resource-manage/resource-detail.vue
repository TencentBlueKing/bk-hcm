<script lang="ts" setup>
import HostDetail from './children/detail/host-detail.vue';
import VpcDetail from './children/detail/vpc-detail.vue';
import SubnetDetail from './children/detail/subnet-detail.vue';
import SecurityDetail from './children/detail/security-detail.vue';
import GcpDetail from './children/detail/gcp-detail.vue';
import DriveDetail from './children/detail/drive-detail.vue';
import IpDetail from './children/detail/ip-detail.vue';
import RoutingDetail from './children/detail/routing-detail.vue';
import ImageDetail from './children/detail/image-detail.vue';
import NetworkInterfaceDetail from './children/detail/network-interface-detail.vue';
import TemplateDetail from './children/detail/template-detail';
import { useVerify } from '@/hooks';
import bus from '@/common/bus';

import { provide, computed } from 'vue';

import { useRoute } from 'vue-router';

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
  eips: IpDetail,
  route: RoutingDetail,
  gcp: GcpDetail,
  image: ImageDetail,
  'network-interface': NetworkInterfaceDetail,
  template: TemplateDetail,
};

const renderComponent = computed(() => {
  return componentMap[route.params.type as string];
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
  <div>
    <component :is="renderComponent"></component>
    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>
  </div>
</template>

<style lang="scss">
.delete-resource-infobox,
.recycle-resource-infobox {
  .bk-info-sub-title {
    word-break: break-all;
  }
}
</style>
