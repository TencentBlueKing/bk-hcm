<script setup lang="ts">
import { computed, onBeforeUnmount } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import HcmBreadcrumb from '@/components/layout/breadcrumb.vue';
import { MENU_RESOURCE_RESOURCE_LIST } from '@/constants/menu-symbol';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import AccountVendorGroup from './account/vendor-group/index.vue';

const route = useRoute();
const resourceAccountStore = useResourceAccountStore();

const isListPage = computed(() => route.name === MENU_RESOURCE_RESOURCE_LIST);

// 列表、新建、详情同在本入口下，互相切换不卸载；离开资源纳管时清空账号选择
onBeforeUnmount(() => resourceAccountStore.clear());
</script>

<template>
  <div class="resource-entry">
    <div class="resource-entry-sidebar">
      <AccountVendorGroup />
    </div>
    <div class="resource-entry-main">
      <HcmBreadcrumb v-if="!isListPage" />
      <div class="resource-entry-content" :class="{ 'is-list-page': isListPage }">
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
  min-width: 0;
  overflow: hidden;
}

.resource-entry-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;

  &.is-list-page {
    overflow: hidden;
  }
}
</style>
