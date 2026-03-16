<script setup lang="ts">
import { computed, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_BUSINESS_RESOURCE_PLAN_CVM, MENU_BUSINESS_RESOURCE_PLAN_GPU } from '@/constants/menu-symbol';
import CvmPrediction from '@/views/business/resource-plan/list';
import GpuDemand from './gpu/list/index.vue';

const router = useRouter();
const route = useRoute();

const tabPanels = [
  { name: MENU_BUSINESS_RESOURCE_PLAN_CVM, label: 'CVM预测' },
  { name: MENU_BUSINESS_RESOURCE_PLAN_GPU, label: 'GPU需求' },
];

const tabActive = computed({
  get() {
    return (route.name as string) || tabPanels[0].name;
  },
  set(value) {
    router.push({ name: value, query: { [GLOBAL_BIZS_KEY]: route.query?.[GLOBAL_BIZS_KEY] } });
  },
});

provide('isBusinessPage', true);

const tabComps: Record<string, any> = {
  [MENU_BUSINESS_RESOURCE_PLAN_CVM]: CvmPrediction,
  [MENU_BUSINESS_RESOURCE_PLAN_GPU]: GpuDemand,
};
</script>

<template>
  <bk-tab class="resource-plan-entry" type="unborder-card" v-model:active="tabActive">
    <bk-tab-panel
      v-for="panel in tabPanels"
      :key="panel.name"
      :name="panel.name"
      :label="panel.label"
      render-directive="'if'"
    >
      <component :is="tabComps[tabActive]" v-if="tabActive === panel.name" />
    </bk-tab-panel>
  </bk-tab>
</template>

<style lang="scss" scoped>
.resource-plan-entry {
  height: 100%;

  :deep(.bk-tab-header) {
    padding: 0 12px;
  }

  :deep(.bk-tab-content) {
    padding: 16px;
  }
}
</style>
