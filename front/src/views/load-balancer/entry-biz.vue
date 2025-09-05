<script setup lang="ts">
import { computed, provide } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { MENU_BUSINESS_RESOURCE_OVERVIEW, MENU_BUSINESS_DEVICE_SEARCH_OVERVIEW } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import {
  AUTH_BIZ_CREATE_CLB,
  AUTH_BIZ_DELETE_CLB,
  AUTH_BIZ_UPDATE_CLB,
  AUTH_CREATE_CLB,
  AUTH_DELETE_CLB,
  AUTH_UPDATE_CLB,
} from '@/constants/auth-symbols';
import { getAuthSignByBusinessId } from '@/utils';
import routerAction from '@/router/utils/action';

import ResourceView from './resource/index.vue';
import DeviceSearchView from './device/index.vue';

const route = useRoute();
const { t } = useI18n();

const LOAD_BALANCER_VIEW_LIST = [
  {
    label: t('资源列表'),
    path: '/business/load-balancer/resource',
    name: MENU_BUSINESS_RESOURCE_OVERVIEW,
    component: ResourceView,
  },
  {
    label: t('设备检索'),
    path: '/business/load-balancer/device',
    name: MENU_BUSINESS_DEVICE_SEARCH_OVERVIEW,
    component: DeviceSearchView,
  },
];

const activeComponent = computed(
  () => LOAD_BALANCER_VIEW_LIST.find((item) => route.path.includes(item.path)).component,
);

const currentGlobalBusinessId = computed(() => {
  const val = route.query?.[GLOBAL_BIZS_KEY];
  return val ? Number(val) : undefined;
});
const clbCreateAuthSign = computed(() =>
  getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_CREATE_CLB, AUTH_BIZ_CREATE_CLB),
);
const clbOperationAuthSign = computed(() =>
  getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_UPDATE_CLB, AUTH_BIZ_UPDATE_CLB),
);
const clbDeleteAuthSign = computed(() =>
  getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_DELETE_CLB, AUTH_BIZ_DELETE_CLB),
);

const handleViewChange = (name: (typeof LOAD_BALANCER_VIEW_LIST)[number]['name']) => {
  routerAction.redirect({
    name,
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
  });
};

provide('currentGlobalBusinessId', currentGlobalBusinessId);
provide('clbCreateAuthSign', clbCreateAuthSign);
provide('clbOperationAuthSign', clbOperationAuthSign);
provide('clbDeleteAuthSign', clbDeleteAuthSign);
</script>

<template>
  <div class="home">
    <div class="header">
      <span class="title">{{ t('负载均衡') }}</span>
      <ul class="view-list">
        <li
          v-for="{ label, name, path } in LOAD_BALANCER_VIEW_LIST"
          :key="name"
          class="view-item"
          :class="{ active: route.path.includes(path) }"
          @click="handleViewChange(name)"
        >
          {{ label }}
        </li>
      </ul>
    </div>
    <div class="main">
      <component :is="activeComponent" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.home {
  height: 100%;
  background-color: #fff;

  .header {
    display: flex;
    position: relative;
    justify-content: center;
    align-items: center;
    height: 52px;
    box-shadow: 0 3px 4px 0 #0000000a;

    .title {
      position: absolute;
      left: 24px;
      font-size: 16px;
      color: #313238;
      line-height: 24px;
    }

    .view-list {
      position: relative;
      left: -8px;
      display: flex;

      .view-item {
        position: relative;
        padding: 0 24px;
        height: 52px;
        line-height: 52px;
        cursor: pointer;

        &.active {
          background-color: #f0f5ff;
          color: #3a84ff;

          &::before {
            position: absolute;
            content: '';
            left: 0;
            width: 100%;
            height: 3px;
            background-color: #3a84ff;
          }
        }
      }
    }
  }

  .main {
    height: calc(100% - 52px);
  }
}
</style>
