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
import routerAction from '@/router/utils/action';
import { MENU_RESOURCE_MANAGE } from '@/constants/menu-symbol';

const route = useRoute();
const accountStore = useAccountStore();

const componentMap = {
  host: HostDetail,
  vpc: VpcDetail,
  subnet: SubnetDetail,
  security: SecurityDetail,
  drive: DriveDetail,
  eip: IpDetail,
  routing: RoutingDetail,
  gcp: GcpDetail,
  image: ImageDetail,
  'network-interface': NetworkInterfaceDetail,
  template: TemplateDetail,
};

// 旧地址的类型名
const resourceTypeToComponentKey: Record<string, string> = {
  ip: 'eip',
  eips: 'eip',
  route: 'routing',
};
const resourceType = computed(() => route.params.resourceType as string);
const renderComponent = computed(() => {
  const componentKey = resourceTypeToComponentKey[resourceType.value] || resourceType.value;
  return componentMap[componentKey as keyof typeof componentMap];
});

const isResourcePage = computed(() => {
  // 资源下没有业务ID
  return !accountStore.bizs;
});

provide('isResourcePage', isResourcePage);

const handleBackToManage = () => {
  routerAction.redirect({ name: MENU_RESOURCE_MANAGE });
};
</script>

<template>
  <!-- 详情间互跳只换 params，按类型 + ID 重建组件，避免沿用旧 ID 的数据 -->
  <component v-if="renderComponent" :is="renderComponent" :key="`${resourceType}-${route.params.id}`"></component>
  <bk-exception v-else type="404" scene="part" description="资源类型不存在">
    <bk-button theme="primary" @click="handleBackToManage">返回资源纳管</bk-button>
  </bk-exception>
</template>

<style lang="scss">
.delete-resource-infobox,
.recycle-resource-infobox {
  .bk-info-sub-title {
    word-break: break-all;
  }
}
</style>
