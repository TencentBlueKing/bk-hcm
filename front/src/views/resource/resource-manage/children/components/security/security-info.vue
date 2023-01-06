<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const settingInfo = [
  {
    name: t('账号'),
    value: '我的账号',
    prop: 'account',
  },
  {
    name: t('资源ID'),
    value: '100',
    prop: 'id',
  },
  {
    name: t('资源名称'),
    value: '100',
    prop: 'name',
  },
  {
    name: t('云厂商'),
    value: '100',
    prop: 'vendor',
  },
  {
    name: t('业务'),
    value: '100',
    prop: 'bk_biz_id',
  },
  {
    name: t('地域 ID'),
    value: 'mirror.tencent',
    prop: 'region',
  },
  {
    name: t('地域'),
    value: 'mirror.tencent',
    prop: 'region',
  },
  {
    name: t('关联CVM实例数'),
    value: '11111',
    prop: 'cvm_ids',
  },
  {
    name: t('创建时间'),
    value: '11111',
    prop: 'create_at',
  },
  {
    name: t('修改时间'),
    value: '11111',
    prop: 'update_at',
  },
  {
    name: t('备注'),
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
);
detail.value = { id: 1, memo: '备注', name: 'test' } || detail;
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-info class="mt20" :fields="settingInfo" :detail="detail"></detail-info>
  </bk-loading>
</template>
