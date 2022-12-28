<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import GcpRelate from '../components/gcp/gcp-relate.vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const gcpFields = [
  {
    name: '资源ID',
    value: '1234223',
    prop: 'id',
  },
  {
    name: '资源名称',
    value: '1234223',
    link: 'http://www.baidu.com',
    prop: 'account-id',
  },
  {
    name: '账号',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '业务',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '云厂商',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '日志',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: 'vpc',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '优先级',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '方向',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '对匹配项执行的操作',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '目标',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '来源过滤条件',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '实施',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '创建时间',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '修改时间',
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: '备注',
    value: '1234223',
    edit: true,
    prop: 'memo',
  },
];


const {
  loading,
  detail,
} = useDetail(
  'cloud/vendors/gcp/firewalls/rules',
  '1',
);
const gcpDetail = { id: 1, memo: '备注' } || detail;

const tabs = [
  {
    name: '关联实例',
    value: 'relate',
  },
];
</script>

<template>
  <detail-header>
    GCP防火墙：ID（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('删除') }}
      </bk-button>
    </template>
  </detail-header>
  <bk-loading
    :loading="loading"
  >
    <detail-info
      :fields="gcpFields"
      :detail="gcpDetail"
    />
  </bk-loading>

  <detail-tab
    :tabs="tabs"
  >
    <gcp-relate />
  </detail-tab>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
