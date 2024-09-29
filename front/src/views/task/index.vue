<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import clb from './clb.vue';

const router = useRouter();
const route = useRoute();

const tabPanels = [{ name: 'clb', label: '负载均衡' }];
const tabActive = computed({
  get() {
    return route.params.resourceType || tabPanels[0].name;
  },
  set(value) {
    router.push({ params: { resourceType: value }, query: route.query });
  },
});

const tabComps: Record<string, any> = { clb };
</script>

<template>
  <bk-tab class="page-task" type="unborder-card" v-model:active="tabActive">
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
.page-task {
  :deep(.bk-tab-header) {
    padding: 0 12px;
  }
  :deep(.bk-tab-content) {
    padding: 16px;
  }
}
</style>
