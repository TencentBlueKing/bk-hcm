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
import { useVerify } from '@/hooks';

import {
  computed,
} from 'vue';
import {
  useRoute,
} from 'vue-router';

const route = useRoute();

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
};

const renderComponent = computed(() => {
  return componentMap[route.params.type as string];
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
