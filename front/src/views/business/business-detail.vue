<script lang="ts" setup>
import {
  computed
} from 'vue';
import {
  useRoute,
} from 'vue-router';

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

const route = useRoute();

const componentMap = {
  host: HostDetail,
  vpc: VpcDetail,
  subnet: SubnetDetail,
  security: SecurityDetail,
  drive: DriveDetail,
  ip: IpDetail,
  route: RoutingDetail,
  gcp: GcpDetail,
  image: ImageDetail,
  'network-interface': NetworkInterfaceDetail,
};

const renderComponent = computed(() => {
  return Object.keys(componentMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = componentMap[cur]
    return acc
  }, {})
})
</script>

<template>
  <component :is="renderComponent"></component>
</template>
