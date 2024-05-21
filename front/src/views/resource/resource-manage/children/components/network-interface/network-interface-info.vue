<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { ref } from 'vue';
import { CloudType } from '@/typings';
import { useRegionsStore } from '@/store/useRegionsStore';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const { getRegionName } = useRegionsStore();

const fields = ref([
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '云资源 ID',
    prop: 'cloud_id',
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: string) {
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
    link(val: string) {
      return `/#/resource/account/detail/?accountId=${val}&id=${val}`;
    },
  },
  {
    name: '资源组',
    prop: 'resource_group_name',
  },
  {
    name: '位置',
    prop: 'region',
    render: (val: string) => getRegionName(props?.detail?.vendor, val),
  },
]);
switch (props?.detail.vendor) {
  case 'azure':
    fields.value.push(
      ...[
        {
          name: '内网IPv4地址',
          prop: 'private_ipv4',
          render(cell: string[]) {
            return cell.length ? cell.join(',') : '--';
          },
        },
        {
          name: '公网IPv4地址',
          prop: 'public_ipv4',
          render(cell: string[]) {
            return cell.length ? cell.join(',') : '--';
          },
        },
        {
          name: '专用IP地址(IPv6)',
          prop: 'private_ipv6',
          render(cell: string[]) {
            return cell.length ? cell.join(',') : '--';
          },
        },
        {
          name: '公用IP地址(IPv6)',
          prop: 'public_ipv6',
          render(cell: string[]) {
            return cell.length ? cell.join(',') : '--';
          },
        },
        {
          name: '更快的网络连接',
          prop: 'enable_accelerated_networking',
          render(val: boolean) {
            return val ? '是' : '否';
          },
        },
        {
          name: '已关联到主机ID',
          prop: 'instance_id',
          render(cell: string) {
            return cell || '--';
          },
        },
        {
          name: '已关联到安全组ID',
          prop: 'cloud_security_group_id',
          render(cell: string) {
            return cell || '--';
          },
        },
        {
          name: '虚拟网络/子网',
          prop: 'virtualNetworkSubnetId',
        },
        {
          name: 'MAC地址',
          prop: 'mac_address',
        },
      ],
    );
    break;
}
</script>

<template>
  <div class="field-list">
    <detail-info :detail="detail" :fields="fields" />
  </div>
</template>
