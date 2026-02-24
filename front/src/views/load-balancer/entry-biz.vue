<script setup lang="ts">
import { computed, provide } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH, MENU_BUSINESS_LOAD_BALANCER_LB_VIEW } from '@/constants/menu-symbol';
import {
  AUTH_BIZ_CREATE_CLB,
  AUTH_BIZ_DELETE_CLB,
  AUTH_BIZ_UPDATE_CLB,
  AUTH_CREATE_CLB,
  AUTH_DELETE_CLB,
  AUTH_UPDATE_CLB,
} from '@/constants/auth-symbols';
import { getAuthSignByBusinessId } from '@/utils';

const route = useRoute();
const { t } = useI18n();

const LOAD_BALANCER_VIEWS = [
  { label: t('资源列表'), name: MENU_BUSINESS_LOAD_BALANCER_LB_VIEW },
  { label: t('配置检索'), name: MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH },
];

const activeView = computed(() =>
  route.name === MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH
    ? MENU_BUSINESS_LOAD_BALANCE_DEVICE_SEARCH
    : MENU_BUSINESS_LOAD_BALANCER_LB_VIEW,
);

const currentGlobalBusinessId = computed(() => {
  const val = route.params.bizId;
  return val ? Number(val as string) : undefined;
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
        <router-link
          v-for="view in LOAD_BALANCER_VIEWS"
          :key="view.name.toString()"
          :to="{ name: view.name as any }"
          custom
          v-slot="{ navigate }"
        >
          <li class="view-item" :class="{ active: activeView === view.name }" @click="navigate">
            {{ view.label }}
          </li>
        </router-link>
      </ul>
    </div>
    <div class="main">
      <router-view />
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
