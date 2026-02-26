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

import { provide, computed } from 'vue';

import { useRoute } from 'vue-router';

import { useAccountStore } from '@/store';

const route = useRoute();
const accountStore = useAccountStore();

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

const resourceTypeToComponentKey: Record<string, string> = {
  ip: 'eips',
  routing: 'route',
};
const renderComponent = computed(() => {
  const resourceType = route.params.resourceType as string;
  const componentKey = resourceTypeToComponentKey[resourceType] || resourceType;
  return componentMap[componentKey as keyof typeof componentMap];
});

const isResourcePage = computed(() => {
  return !accountStore.bizs;
});

provide('isResourcePage', isResourcePage);
</script>

<template>
  <div>
    <component :is="renderComponent"></component>
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
