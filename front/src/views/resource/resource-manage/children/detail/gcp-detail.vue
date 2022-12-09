<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import GcpRelate from '../components/gcp/gcp-relate.vue';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const gcpFields = [
  {
    name: 'ID',
    value: '1234223',
  },
  {
    name: '账号ID',
    value: '1234223',
    link: 'http://www.baidu.com',
  },
  {
    name: '账号名称',
    value: '1234223',
  },
  {
    name: '备注',
    value: '1234223',
    edit: true,
  },
  {
    name: 'VPC',
    value: '1234223',
    copy: '1234223',
  },
];
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

  <detail-info
    :fields="gcpFields"
  />

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
