<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref, watchEffect } from 'vue';
import { CloudType } from '@/typings';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';

const props = defineProps({
  detail: {
    type: Object,
  },
  isResourcePage: {
    type: Boolean,
  },
});

const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();

const fields = ref([
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
    name: '所属VPC',
    prop: 'cloud_vpc_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
  },
  {
    name: '已关联到',
    prop: 'cvm_id',
    link(val: string) {
      if (props.isResourcePage) {
        return val ? `/#/resource/detail/host?id=${val}&type=gcp` : '--';
      }
      return val ? `/#/business/host/detail?id=${val}&type=gcp&bizs=${props.detail.bk_biz_id}` : '--';
    },
  },
  {
    name: '网络层级',
    prop: 'networkTier',
    render(val: string) {
      const vals = { PREMIUM: '高级', STANDARD: '标准' };
      return vals[val];
    },
  },
  {
    name: 'IP转发',
    prop: 'can_ip_forward',
    render(val: boolean) {
      return val ? '开启' : '关闭';
    },
  },
]);

const data = ref([]);

watchEffect(() => {
  data.value = {
    ...props.detail,
    networkTier: props.detail?.access_configs?.[0]?.network_tier,
  };
});
</script>

<template>
  <div class="field-list">
    <detail-info :detail="data" :fields="fields" />
  </div>
</template>
