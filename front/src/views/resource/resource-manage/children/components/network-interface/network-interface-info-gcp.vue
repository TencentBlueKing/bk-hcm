<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref } from 'vue';
import { useRoute } from 'vue-router';
import { CloudType } from '@/typings';
import { formatGcpNetworkTier } from '../../../common/gcp-network-tier';
import routerAction from '@/router/utils/action';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  MENU_BUSINESS_HOST_DETAILS,
  MENU_RESOURCE_ACCOUNT_DETAIL,
  MENU_RESOURCE_DETAIL,
} from '@/constants/menu-symbol';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { FieldList } from '../../../common/info-list/types';

const props = defineProps<{ detail?: VendorResource<NetworkInterface, 'gcp'>; isResourcePage?: boolean }>();

const fieldPath = extensionPath<VendorResource<NetworkInterface, 'gcp'>>();

const route = useRoute();
const { getBizsId } = useWhereAmI();
const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();

const fields = ref<FieldList>([
  {
    name: '资源ID',
    prop: 'id',
  },
  {
    name: '云资源ID',
    prop: 'cloud_id',
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: keyof typeof CloudType) {
      return CloudType[cell] || '--';
    },
  },
  {
    name: '网络接口名称',
    prop: 'name',
  },
  {
    name: '账号',
    prop: 'account_id',
    txtBtn(accountId: string) {
      routerAction.redirect({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId } }, { history: true });
    },
  },
  {
    name: '地域',
    prop: 'region',
    render: (val: string) => getRegionName(props?.detail?.vendor, val),
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
  },
  {
    name: '内网IPv4地址',
    prop: 'private_ipv4',
    render(cell: string[]) {
      return cell?.join(',') || '--';
    },
  },
  {
    name: '公网IPv4地址',
    prop: 'public_ipv4',
    render(cell: string[]) {
      return cell?.join(',') || '--';
    },
  },
  {
    name: '专用IP地址(IPv6)',
    prop: 'private_ipv6',
    render(cell: string[]) {
      return cell?.join(',') || '--';
    },
  },
  {
    name: '公用IP地址(IPv6)',
    prop: 'public_ipv6',
    render(cell: string[]) {
      return cell?.join(',') || '--';
    },
  },
  {
    name: '所属VPC',
    prop: 'cloud_vpc_id',
    render(val: string) {
      if (!val) return '--';
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
    copyContent: (cell) => cell,
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_id',
    render(val: string) {
      if (!val) return '--';
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
    copyContent: (cell) => cell,
  },
  {
    name: '已关联到',
    prop: 'cvm_id',
    txtBtn(id: string) {
      const routeInfo = props.isResourcePage
        ? {
            name: MENU_RESOURCE_DETAIL,
            params: { resourceType: 'host', id },
            query: { accountId: route.query.accountId, vendor: props.detail.vendor },
          }
        : {
            name: MENU_BUSINESS_HOST_DETAILS,
            params: { bizId: getBizsId(), id },
            query: { vendor: props.detail.vendor },
          };
      routerAction.redirect(routeInfo, { history: true });
    },
  },
  {
    name: '网络层级',
    prop: fieldPath('extension.access_configs[0].network_tier'),
    render: formatGcpNetworkTier,
  },
  {
    name: 'IP转发',
    prop: fieldPath('extension'),
    render(extension: VendorResource<NetworkInterface, 'gcp'>['extension']) {
      if (!extension) return '--';
      return extension.can_ip_forward ? '开启' : '关闭';
    },
  },
]);
</script>

<template>
  <div class="field-list">
    <detail-info :detail="props.detail" :fields="fields" global-copyable label-width="150px" />
  </div>
</template>
