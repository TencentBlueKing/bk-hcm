<script lang="ts" setup>
import InfoList from '../../../common/info-list/info-list';

import useDetail from '../../../hooks/use-detail';

const settingFields = [
  {
    name: '账号',
    prop: 'account_id',
    link: 'http://www.baidu.com',
  },
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '资源名称',
    prop: 'name',
  },
  {
    name: '云厂商',
    prop: 'vendor',
  },
  {
    name: '业务',
    prop: '1234223',
  },
  {
    name: '状态',
    prop: 'status',
  },
  {
    name: '是否默认子网',
    prop: 'is_default',
  },
  {
    name: '所属 VPC',
    prop: 'is_default',
  },
  {
    name: 'IPv4 CIDR',
    prop: 'ipv4_cidr',
  },
  {
    name: '可用 IPv4 地址数',
    prop: 'ipv4_cidr',
  },
  {
    name: 'IPv6 CIDR',
    prop: 'ipv6_cidr',
  },
  {
    name: '可用 IPv6 地址数',
    prop: 'ipv4_cidr',
  },
  {
    name: '地域 ID',
    prop: 'region',
  },
  {
    name: '地域名称',
    prop: '12312321',
  },
  {
    name: '可用区 ID',
    prop: '1234223',
  },
  {
    name: '可用区名称',
    prop: '1234223',
  },
  {
    name: '创建时间',
    prop: 'create_at',
  },
  {
    name: '备注',
    prop: 'memo',
    edit: true,
  },
];

const {
  loading,
  detail,
} = useDetail(
  'subnet',
  '1',
  settingFields,
);
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <info-list class="mt20" :fields="detail"></info-list>
  </bk-loading>
</template>
