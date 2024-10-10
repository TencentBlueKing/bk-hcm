<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import RouteInfo from '../components/route/route-info.vue';
import RouteSubnet from '../components/route/route-subnet.vue';

import { useRoute } from 'vue-router';
import useDetail from '../../hooks/use-detail';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const route = useRoute();
const { whereAmI } = useWhereAmI();

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

const { loading, detail } = useDetail('route_tables', route.query.id as string);
</script>

<template>
  <bk-loading :loading="loading">
    <detail-header>路由表：ID（{{ detail.id }}）</detail-header>
    <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
      <detail-tab :tabs="routeTabs" class="route-tab">
        <template #default="type">
          <route-info v-if="type === 'detail'" :detail="detail" />
          <route-subnet v-if="type === 'subnet'" :detail="detail" />
        </template>
      </detail-tab>
    </div>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}

.w60 {
  width: 60px;
}
</style>
