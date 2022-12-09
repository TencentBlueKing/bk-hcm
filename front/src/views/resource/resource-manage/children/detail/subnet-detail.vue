<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import SubnetInfo from '../../children/components/subnet/subnet-info.vue';
import SubnetRoute from '../../children/components/subnet/subnet-route.vue';

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
  {
    name: '路由策略',
    value: 'network',
  },
];
</script>

<template>
  <detail-header>
    子网：ID（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('删除') }}
      </bk-button>
    </template>
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default="type">
      <subnet-info v-if="type === 'detail'"></subnet-info>
      <subnet-route v-if="type === 'network'"></subnet-route>
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
