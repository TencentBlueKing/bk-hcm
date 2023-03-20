<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { ImageTypeEnum } from '@/typings';

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
    prop: 'cloud_id',
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
    render() {
      return props.data.zones;
    },
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
  },
  {
    name: '创建时间',
    prop: 'cloud_created_time',
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
    name: '接口名称',
    prop: 'cloud_network_interface_ids',
  },
  {
    name: '私有IPv4地址',
    prop: 'private_ipv4_addresses',
  },
  {
    name: '公有IPv4地址',
    prop: 'public_ipv4_addresses',
  },
  {
    name: '私有IPv6地址',
    prop: 'private_ipv6_addresses',
  },
  {
    name: '公有IPv6地址',
    prop: 'public_ipv6_addresses',
  },
];

const settingInfo = [
  {
    name: '实例规格',
    prop: 'machine_type',
  },
  {
    name: 'CPU',
    prop: 'cpu',
  },
  {
    name: '内存',
    prop: 'memory',
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
