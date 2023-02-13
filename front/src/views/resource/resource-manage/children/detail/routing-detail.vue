<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import RouteInfo from '../components/route/route-info.vue';
import RouteSubnet from '../components/route/route-subnet.vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRoute,
} from 'vue-router';
import useDetail from '../../hooks/use-detail';
const route = useRoute();

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

const {
  t,
} = useI18n();

const {
  loading,
  detail,
} = useDetail(
  'route_tables',
  route.query.id as string,
);
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      路由表：（{{ detail.id }}）
    </detail-header>

    <detail-tab
      :tabs="routeTabs"
      class="route-tab"
    >
      <template #default="type">
        <route-info v-if="type === 'detail'" :detail="detail" />
        <route-subnet v-if="type === 'subnet'" :detail="detail" />
      </template>
    </detail-tab>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
:deep(.detail-tab-main) .bk-tab-content {
  height: calc(100vh - 300px);
}
</style>
