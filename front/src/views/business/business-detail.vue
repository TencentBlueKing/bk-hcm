<script lang="ts" setup>
import { provide, computed, type Component } from 'vue';
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

import { useAccountStore } from '@/store';
import {
  MENU_BUSINESS_HOST_DETAILS,
  MENU_BUSINESS_VPC_DETAILS,
  MENU_BUSINESS_SUBNET_DETAILS,
  MENU_BUSINESS_SECURITY_GROUP_DETAILS,
  MENU_BUSINESS_DRIVE_DETAILS,
  MENU_BUSINESS_EIP_DETAILS,
  MENU_BUSINESS_ROUTE_TABLE_DETAILS,
  MENU_BUSINESS_GCP_DETAILS,
  MENU_BUSINESS_IMAGE_DETAILS,
  MENU_BUSINESS_NIF_DETAILS,
  MENU_BUSINESS_TEMPLATE_DETAILS,
} from '@/constants/menu-symbol';

const route = useRoute();
const accountStore = useAccountStore();

// 详情 ID 在路径里，按路由 name 取组件（按路径子串匹配会被 ID 干扰）
const componentMap = new Map<symbol, Component>([
  [MENU_BUSINESS_HOST_DETAILS, HostDetail],
  [MENU_BUSINESS_VPC_DETAILS, VpcDetail],
  [MENU_BUSINESS_SUBNET_DETAILS, SubnetDetail],
  [MENU_BUSINESS_SECURITY_GROUP_DETAILS, SecurityDetail],
  [MENU_BUSINESS_DRIVE_DETAILS, DriveDetail],
  [MENU_BUSINESS_EIP_DETAILS, IpDetail],
  [MENU_BUSINESS_ROUTE_TABLE_DETAILS, RoutingDetail],
  [MENU_BUSINESS_GCP_DETAILS, GcpDetail],
  [MENU_BUSINESS_IMAGE_DETAILS, ImageDetail],
  [MENU_BUSINESS_NIF_DETAILS, NetworkInterfaceDetail],
  [MENU_BUSINESS_TEMPLATE_DETAILS, TemplateDetail],
]);

const renderComponent = computed(() => componentMap.get(route.name as symbol));

const isResourcePage = computed(() => {
  return !accountStore.bizs;
});

provide('isResourcePage', isResourcePage);
</script>

<template>
  <!-- 详情间互跳只换 params，按 ID 重建组件，避免沿用旧 ID 的数据 -->
  <component :is="renderComponent" :key="route.params.id"></component>
</template>
