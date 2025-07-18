<script setup lang="ts">
import { computed, provide, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { MENU_BUSINESS_LOAD_BALANCER_OVERVIEW, MENU_BUSINESS_TARGET_GROUP_OVERVIEW } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';

import LoadBalancerView from './clb/index.vue';
import TargetGroupView from './target-group/index.vue';

const route = useRoute();
const { t } = useI18n();

const LOAD_BALANCER_VIEW_LIST = [
  {
    label: t('负载均衡视角'),
    path: '/business/load-balancer/clb',
    name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
    component: LoadBalancerView,
  },
  {
    label: t('目标组视角'),
    path: '/business/load-balancer/target-group',
    name: MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
    component: TargetGroupView,
  },
];

const activeComponent = computed(
  () => LOAD_BALANCER_VIEW_LIST.find((item) => route.path.includes(item.path)).component,
);

const currentGlobalBusinessId = computed(() => {
  const val = route.query?.[GLOBAL_BIZS_KEY];
  return val ? Number(val) : undefined;
});

const handleViewChange = (name: (typeof LOAD_BALANCER_VIEW_LIST)[number]['name']) => {
  routerAction.redirect({
    name,
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
  });
};

provide('currentGlobalBusinessId', currentGlobalBusinessId);

// TODO-CLB: createClbActionName、deleteClbActionName是为了兼容旧版负载均衡的逻辑
const createClbActionName = ref('biz_clb_resource_create');
const deleteClbActionName = ref('biz_clb_resource_delete');
provide('createClbActionName', createClbActionName);
provide('deleteClbActionName', deleteClbActionName);
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
