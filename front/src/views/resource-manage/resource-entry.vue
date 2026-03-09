<script setup lang="ts">
/**
 * 资源纳管入口组件
 * - 左侧：AccountVendorGroup 云厂商&账号选择器
 * - 右侧：详情/申请页显示面包屑 + RouterView
 */
import { computed } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import AccountVendorGroup from './account/vendor-group/index.vue';
import HcmBreadcrumb from '@/components/layout/breadcrumb.vue';
import { MENU_RESOURCE_RESOURCE_LIST } from '@/constants/menu-symbol';

const route = useRoute();
const isListPage = computed(() => route.name === MENU_RESOURCE_RESOURCE_LIST);
</script>

<template>
  <div class="resource-entry">
    <div class="resource-entry-sidebar">
      <AccountVendorGroup />
    </div>
    <div class="resource-entry-main">
      <HcmBreadcrumb v-if="!isListPage" />
      <div class="resource-entry-content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.resource-entry {
  display: flex;
  height: 100%;
}

.resource-entry-sidebar {
  flex-shrink: 0;
  width: 240px;
  height: 100%;
  border-right: 1px solid #eaebf0;
  overflow-y: auto;
}

.resource-entry-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.resource-entry-content {
  flex: 1;
  overflow-y: auto;
}
</style>
