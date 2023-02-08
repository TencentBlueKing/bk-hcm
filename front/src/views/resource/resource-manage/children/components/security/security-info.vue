<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

import {
  useI18n,
} from 'vue-i18n';

import {
  PropType,
} from 'vue';

const props = defineProps({
  id: {
    type: String as PropType<any>,
  },
});

const {
  t,
} = useI18n();

const settingInfo = [
  {
    name: 'ID',
    value: '100',
    prop: 'id',
  },
  {
    name: t('账号 ID'),
    value: '我的账号',
    prop: 'account_id',
  },
  {
    name: t('资源名称'),
    value: '100',
    prop: 'name',
  },
  {
    name: t('云资源ID'),
    value: '100',
    prop: 'cloud_id',
  },
  {
    name: t('云厂商'),
    value: '100',
    prop: 'vendorName',
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
    prop: 'cvm_count',
  },
  {
    name: t('创建时间'),
    value: '11111',
    prop: 'created_at',
  },
  {
    name: t('修改时间'),
    value: '11111',
    prop: 'updated_at',
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
  'security_groups',
  props.id,
);
console.log('detail', detail.value, loading);
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-info class="mt20" :fields="settingInfo" :detail="detail"></detail-info>
  </bk-loading>
</template>
