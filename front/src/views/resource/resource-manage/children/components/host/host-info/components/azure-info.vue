<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import { PropType } from 'vue';
import { TypeEnum, useRouteLinkBtn } from '@/hooks/useRouteLinkBtn';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import { timeFormatter } from '@/common/util';

const shortCutStr = (str: string) => {
  if (!str) return str;
  return str.split('/').reverse()[0];
};

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
    render: () => shortCutStr(props.data.cloud_id),
  },
  {
    name: '账号',
    prop: 'account_id',
    render: () =>
      useRouteLinkBtn(props.data, {
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
    render() {
      return props.data.bk_biz_id === -1 ? '未分配' : `${props.data.bk_biz_id_name} (${props.data.bk_biz_id})`;
    },
  },
  {
    name: '创建时间',
    prop: 'cloud_created_time',
    render() {
      return timeFormatter(props.data.cloud_created_time);
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
    name: '所属网络',
    prop: 'cloud_vpc_ids',
    render: () =>
      useRouteLinkBtn(props.data, {
        id: 'vpc_ids',
        name: 'cloud_vpc_ids',
        type: TypeEnum.VPC,
      }),
  },
  {
    name: '接口名称',
    prop: 'cloud_network_interface_ids',
    render: () => shortCutStr(props.data.cloud_network_interface_ids[0]),
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
  {
    name: '镜像ID',
    prop: 'cloud_image_id',
    render: () =>
      useRouteLinkBtn(props.data, {
        id: 'image_id',
        type: TypeEnum.IMAGE,
        name: 'cloud_image_id',
      }),
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
    <detail-info :fields="netInfo" :detail="props.data"></detail-info>
  </div>
  <h3 class="info-title">配置信息</h3>
  <div class="wrap-info">
    <detail-info :fields="settingInfo" :detail="props.data"></detail-info>
  </div>
</template>
