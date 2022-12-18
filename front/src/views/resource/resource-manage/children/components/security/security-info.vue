<script lang="ts" setup>
import InfoList from '../../../common/info-list/info-list';

import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

const settingInfo = [
  {
    name: '账号',
    value: '我的账号',
    prop: 'account',
  },
  {
    name: '资源ID',
    value: '100',
    prop: 'resource_id',
  },
  {
    name: '资源名称',
    value: '100',
    prop: 'resource_name',
  },
  {
    name: '业务',
    value: '100',
    prop: 'bk_biz_name',
  },
  {
    name: '地域',
    value: 'mirror.tencent',
    prop: 'region',
  },
  {
    name: '所属VPC',
    value: '11111',
    prop: 'cloud_vpc_id',
  },
  {
    name: '关联CVM实例数',
    value: '11111',
    prop: 'cvm_ids',
  },
  {
    name: '关联模版',
    value: '11111',
    prop: 'cloud_vpc_id',
  },
  {
    name: '资源组',
    value: '11111',
    prop: 'cloud_vpc_id',
  },
  {
    name: '创建时间',
    value: '11111',
    prop: 'create_at',
  },
  {
    name: '修改时间',
    value: '11111',
    prop: 'update_at',
  },
  {
    name: '备注',
    value: '11111',
    prop: 'memo',
  },
];

const {
  loading,
  detail,
} = useDetail(
  'security_group',
  '1',
  settingInfo,
);
detail.value = [
  { name: '账号', value: '我的账号' },
];
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <info-list class="mt20" :fields="detail"></info-list>
  </bk-loading>
</template>
