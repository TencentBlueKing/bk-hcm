<script setup lang="ts">
import { computed, provide, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ActiveQueryKey } from './constants';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';

import LoadBalancerView from './clb/index.vue';
import TargetGroupView from './target-group/index.vue';

const route = useRoute();
const { t } = useI18n();

const LOAD_BALANCER_VIEW_LIST = [
  {
    label: t('负载均衡视角'),
    value: 'load-balancer-view',
    component: LoadBalancerView,
  },
  {
    label: t('目标组视角'),
    value: 'target-group-view',
    component: TargetGroupView,
  },
];

const activeView = ref(route.query?.[ActiveQueryKey.ENTRY] || LOAD_BALANCER_VIEW_LIST[0].value);
const activeComponent = computed(
  () => LOAD_BALANCER_VIEW_LIST.find((item) => item.value === activeView.value).component,
);
const currentGlobalBusinessId = computed(() => {
  const val = route.query?.[GLOBAL_BIZS_KEY];
  return val ? Number(val) : undefined;
});

const handleViewChange = (viewValue: (typeof LOAD_BALANCER_VIEW_LIST)[number]['value']) => {
  activeView.value = viewValue;
  routerAction.redirect({
    query: { [ActiveQueryKey.ENTRY]: viewValue, [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
  });
};

provide('currentGlobalBusinessId', currentGlobalBusinessId);
</script>

<template>
  <div class="home">
    <div class="header">
      <span class="title">{{ t('负载均衡') }}</span>
      <ul class="view-list">
        <li
          v-for="{ label, value } in LOAD_BALANCER_VIEW_LIST"
          :key="value"
          class="view-item"
          :class="{ active: activeView === value }"
          @click="handleViewChange(value)"
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
