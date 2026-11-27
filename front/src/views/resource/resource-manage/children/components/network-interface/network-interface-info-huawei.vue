<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref } from 'vue';
import { CloudType } from '@/typings';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { FieldList } from '../../../common/info-list/types';
import routerAction from '@/router/utils/action';
import { MENU_RESOURCE_ACCOUNT_DETAIL } from '@/constants/menu-symbol';

const props = defineProps<{ detail?: VendorResource<NetworkInterface, 'huawei'> }>();

const fieldPath = extensionPath<VendorResource<NetworkInterface, 'huawei'>>();

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
    name: '业务',
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
  },
  {
    name: '状态',
    prop: fieldPath('extension.port_state'),
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
    name: '已关联到主机ID',
    prop: 'instance_id',
  },
  {
    name: '安全组ID',
    prop: fieldPath('extension.cloud_security_group_ids'), // cloud_security_group_ids
    render(cell: string) {
      return cell || '--';
    },
  },
  {
    name: 'MAC地址',
    prop: fieldPath('extension.mac_addr'),
  },
]);
</script>

<template>
  <div class="field-list">
    <detail-info :detail="props.detail" :fields="fields" global-copyable label-width="150px" />
  </div>
</template>
