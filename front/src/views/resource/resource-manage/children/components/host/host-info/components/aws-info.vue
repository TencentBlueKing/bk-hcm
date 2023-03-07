<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import {
  PropType,
} from 'vue';

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
    prop: 'id',
  },
  {
    name: '账号',
    prop: 'account_id',
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '地域',
    prop: 'region',
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
  },
  {
    name: '启动时间',
    prop: 'cloud_launched_time',
  },
  {
    name: '当前状态',
    prop: 'status',
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
    name: '所属网络',
    prop: 'cloud_vpc_ids',
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_ids',
  },
  // {
  //   name: '用作公网网关',
  //   prop: 'account_id',
  // },
  // {
  //   name: '公网宽带',
  //   prop: 'vendorName',
  // },
  {
    name: '私有 IPv4 地址',
    prop: 'private_ipv4_addresses',
  },
  {
    name: '公有 IPv4 地址',
    prop: 'public_ipv4_addresses',
  },
  {
    name: '私有 IPv6 地址',
    prop: 'private_ipv6_addresses',
  },
  {
    name: '公有IPv6 地址',
    prop: 'public_ipv6_addresses',
  },
  {
    name: '私有 IP DNS 名称(仅限 IPv4)',
    prop: 'private_dns_name',
  },
  {
    name: '公有 IPv4 DNS',
    prop: 'private_dns_name',
  },
];

const settingInfo = [
  {
    name: '实例规格',
    prop: 'machine_type',
  },
  {
    name: 'CPU',
    render() {
      return props.data.cpu_options.core_count;
    },
  },
  {
    name: '操作系统',
    prop: 'os_name',
  },
  {
    name: '镜像id',
    prop: 'cloud_image_id',
  },
];
</script>

<template>
  <h3 class="info-title">实例信息</h3>
  <div class="warp-info">
    <detail-info :fields="cvmInfo" :detail="props.data"></detail-info>
  </div>
  <h3 class="info-title">网络信息</h3>
  <div class="warp-info">
    <detail-info class="mt20" :fields="netInfo" :detail="props.data"></detail-info>
  </div>
  <h3 class="info-title">配置信息</h3>
  <div class="warp-info">
    <detail-info class="mt20" :fields="settingInfo" :detail="props.data"></detail-info>
  </div>
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
  :deep(.host-info) .detail-info-main {
    height: auto !important;
  }
</style>
