<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import RouteInfo from '../components/route/route-info.vue';
import RouteSubnet from '../components/route/route-subnet.vue';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const routeTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
  {
    name: '关联子网',
    value: 'subnet',
  },
];
</script>

<template>
  <detail-header>
    路由表：ID（xxx）
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
    :tabs="routeTabs"
  >
    <template #default="type">
      <route-info v-if="type === 'detail'" />
      <route-subnet v-if="type === 'subnet'" />
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
