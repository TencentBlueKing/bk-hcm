<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import { PropType } from 'vue';
import { TypeEnum, useRouteLinkBtn } from '@/hooks/useRouteLinkBtn';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';
import { timeFormatter } from '@/common/util';

const { getRegionName } = useRegionsStore();

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
});

const cvmInfo = [
  {
    name: '实例名称',
    prop: 'name',
  },
  {
    name: '实例ID',
    prop: 'cloud_id',
  },
  {
    name: '账号',
    prop: 'account_id',
    render: () => useRouteLinkBtn(props.data, {
      id: 'account_id',
      name: 'account_id',
      type: TypeEnum.ACCOUNT,
    }),
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '地域',
    prop: 'region',
    render: () => getRegionName(VendorEnum.GCP, props.data.region),
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render() {
      return props.data.bk_biz_id === -1
        ? '未分配'
        : `${props.data.bk_biz_id_name} (${props.data.bk_biz_id})`;
    },
  },
  {
    name: '启动时间',
    prop: 'cloud_launched_time',
    render() {
      return timeFormatter(props.data.cloud_launched_time);
    },
  },
  {
    name: '当前状态',
    prop: 'status',
    render() {
      return CLOUD_HOST_STATUS[props.data.status];
    },
  },
  {
    name: '实例销毁保护',
    prop: 'disable_api_termination',
    render() {
      return props.data.disable_api_termination ? '是' : '否';
    },
  },
  {
    name: '备注',
    prop: 'memo',
  },
];

const netInfo = [
  {
    name: '接口名称',
    prop: 'cloud_network_interface_ids',
  },
  {
    name: '私有IPv4地址',
    prop: 'private_ipv4_addresses',
    render: (val: string[]) => (val.length ? [...val].join(',') : '--'),
  },
  {
    name: '公有IPv4地址',
    prop: 'public_ipv4_addresses',
    render: (val: string[]) => (val.length ? [...val].join(',') : '--'),
  },
  {
    name: '私有IPv6地址',
    prop: 'private_ipv6_addresses',
    render: (val: string[]) => (val.length ? [...val].join(',') : '--'),
  },
  {
    name: '公有IPv6地址',
    prop: 'public_ipv6_addresses',
    render: (val: string[]) => (val.length ? [...val].join(',') : '--'),
  },
  {
    name: '所属网络',
    prop: 'cloud_vpc_ids',
    render: () => useRouteLinkBtn(props.data, {
      id: 'vpc_ids',
      name: 'cloud_vpc_ids',
      type: TypeEnum.VPC,
    }),
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_ids',
    render: () => useRouteLinkBtn(props.data, {
      id: 'subnet_ids',
      name: 'cloud_subnet_ids',
      type: TypeEnum.SUBNET,
    }),
  },
  // {
  //   name: '网络层级',
  //   prop: 'private_dns_name',
  // },
  {
    name: 'IP转发',
    render() {
      return props.data.can_ip_forward ? '开启' : '关闭';
    },
  },
];

const settingInfo = [
  {
    name: '实例规格',
    prop: 'machine_type',
  },
  {
    name: '操作系统',
    prop: 'os_name',
  },
];
</script>

<template>
  <h3 class="info-title">实例信息</h3>
  <div class="wrap-info">
    <detail-info :fields="cvmInfo" :detail="props.data"></detail-info>
  </div>
  <h3 class="info-title">网络信息</h3>
  <div class="wrap-info">
    <detail-info
      :fields="netInfo"
      :detail="props.data"
    ></detail-info>
  </div>
  <h3 class="info-title">配置信息</h3>
  <div class="wrap-info">
    <detail-info
      :fields="settingInfo"
      :detail="props.data"
    ></detail-info>
  </div>
</template>
