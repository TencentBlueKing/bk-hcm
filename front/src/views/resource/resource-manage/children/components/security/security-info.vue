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
    prop: 'id',
  },
  {
    name: '资源名称',
    value: '100',
    prop: 'name',
  },
  {
    name: '云厂商',
    value: '100',
    prop: 'vendor',
  },
  {
    name: '业务',
    value: '100',
    prop: 'bk_biz_id',
  },
  {
    name: '地域ID',
    value: 'mirror.tencent',
    prop: 'region',
  },
  {
    name: '地域名称',
    value: 'mirror.tencent',
    prop: 'region',
  },
  {
    name: '关联CVM实例数',
    value: '11111',
    prop: 'cvm_ids',
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
  { name: '资源ID', value: 'testid' },
  { name: '备注', value: '1111' },
  { name: '地域ID', value: '地域ID' },
  { name: '资源ID', value: 'testid' },
  { name: '备注', value: '1111' },
];
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <info-list class="mt20" :fields="detail"></info-list>
  </bk-loading>
</template>
