<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import IpInfo from '../components/ip/ip-info.vue';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();


const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <detail-header>
    弹性IP：ID（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('绑定') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('解绑') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('释放') }}
      </bk-button>
    </template>
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default>
      <ip-info></ip-info>
    </template>
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
