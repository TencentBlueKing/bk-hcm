<script setup lang="ts">
import { computed, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { MENU_SERVICE_RESOURCE_PLAN_CVM, MENU_SERVICE_RESOURCE_PLAN_GPU } from '@/constants/menu-symbol';
import { useVerify } from '@/hooks/useVerify';
import CvmPrediction from '@/views/service/resource-plan/resource-manage/list';
import GpuDemand from './gpu/list/index.vue';

const router = useRouter();
const route = useRoute();
const { authVerifyData } = useVerify();

const allTabPanels = [
  { name: MENU_SERVICE_RESOURCE_PLAN_CVM, label: 'CVM预测', auth: 'ziyan_resource_plan_manage' },
  { name: MENU_SERVICE_RESOURCE_PLAN_GPU, label: 'GPU需求', auth: 'ziyan_resource_plan_gpu_demands' },
];

const tabPanels = computed(() =>
  allTabPanels.filter((panel) => !panel.auth || authVerifyData.value?.permissionAction?.[panel.auth]),
);

const tabActive = computed({
  get() {
    const current = route.name as string;
    if (tabPanels.value.some((p) => p.name === current)) return current;
    return tabPanels.value[0]?.name ?? '';
  },
  set(value) {
    router.push({ name: value });
  },
});

provide('isServicePage', true);

const tabComps: Record<string, any> = {
  [MENU_SERVICE_RESOURCE_PLAN_CVM]: CvmPrediction,
  [MENU_SERVICE_RESOURCE_PLAN_GPU]: GpuDemand,
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
    background-color: #fff;
    border-bottom: none;
  }

  :deep(.bk-tab-content) {
    padding: 16px;
  }
}
</style>
