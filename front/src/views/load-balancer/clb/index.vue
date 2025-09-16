<script setup lang="ts">
import { ref, watch, useTemplateRef, inject, Ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import routerAction from '@/router/utils/action';

import LoadBalancerList from '@/views/load-balancer/clb/load-balancer-list.vue';
import TargetGroupList from '@/views/business/load-balancer/group-view/target-group-list';

import { MENU_BUSINESS_LOAD_BALANCER_OVERVIEW, MENU_BUSINESS_TARGET_GROUP_OVERVIEW } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

defineOptions({ name: 'resource-view' });

const route = useRoute();
const { t } = useI18n();
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const typeEnum = {
  clb: {
    value: t('负载均衡'),
    name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
  },
  target_group: {
    value: t('目标组'),
    name: MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
  },
};
// 负载均衡/目标组上次定位的参数
const memory: { [key: string]: any } = {
  clb: {
    name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW,
    query: {},
    params: {},
  },
  target_group: {
    name: MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
    query: {},
    params: {},
  },
};
const type = ref(route?.meta?.extra?.type ?? 'clb');
// TODO-CLB：这里存在一个定位问题（url访问详情页没法定位。可能是virtual-render刚挂载的时候，它的高度计算有问题，第一次的滚动会失效）
const loadBalancerListRef = useTemplateRef<typeof LoadBalancerList>('load-balancer-list');
const handleDetailsShow = (id: string) => {
  loadBalancerListRef.value?.fixToActive(id);
};

// 记录这次的定位位置
const setMemory = (type: string) => {
  const { query, params, name } = route;
  memory[type] = {
    query,
    name,
    params,
  };
};
const getMemory = (type: string) => {
  const { query, params, name } = memory[type];
  return { query, name, params };
};

watch(
  () => type.value,
  (val, oldVal) => {
    setMemory(oldVal);
    const { query, params, name } = getMemory(val);
    routerAction.redirect(
      {
        name,
        query: { ...(query ?? route.query), [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
        params: { ...(params ?? route.params) },
      },
      {
        history: true,
      },
    );
  },
);
</script>

<template>
  <bk-resize-layout class="container" collapsible :initial-divide="300" :min="300">
    <template #aside>
      <bk-radio-group v-model="type" class="resource-type">
        <bk-radio-button label="clb">{{ typeEnum.clb.value }}</bk-radio-button>
        <bk-radio-button label="target_group">{{ typeEnum.target_group.value }}</bk-radio-button>
      </bk-radio-group>
      <target-group-list v-if="type === 'target_group'" />
      <load-balancer-list ref="load-balancer-list" v-else />
    </template>
    <template #main>
      <router-view @details-show="handleDetailsShow" />
    </template>
  </bk-resize-layout>
</template>

<style scoped lang="scss">
.container {
  height: 100%;
}
.resource-type {
  display: flex;
  border-color: #f0f1f5;
  color: #4d4f56;
  margin: 12px 24px 0;
  padding: 3px;
  width: calc(100% - 48px);
  background: #f0f1f5;

  label {
    background: #f0f1f5;
    flex: 1;
    :deep(.bk-radio-button-label) {
      color: #4d4f56;
      border-color: transparent !important;
    }

    &.is-checked {
      :deep(.bk-radio-button-label) {
        background: white;
        font-weight: bold;
      }
    }
  }
}
</style>
