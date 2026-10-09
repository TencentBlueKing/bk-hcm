<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { MENU_RESOURCE } from '@/constants/menu-symbol';
import AccountVendorGroup from '@/views/resource/resource-manage/account/vendor-group/index.vue';
import HcmBreadcrumb from './breadcrumb.vue';

defineProps<{ isRouterAlive: boolean }>();

const route = useRoute();
const breadcrumb = useBreadcrumb();

// TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除
const isLegacyResourceAccess = computed(() => route.matched[0]?.name === MENU_RESOURCE);
const isStatusView = computed(() => route.name === '404' || route.name === 'error');
const showBreadcrumb = computed(() => !isLegacyResourceAccess.value && breadcrumb.data.display);
const isNoPadding = computed(() => route.path === '/service/ticket');
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div :class="['main-view', 'view-warp', { 'no-padding': isNoPadding, 'status-view': isStatusView }]">
      <template v-if="isRouterAlive">
        <div v-if="isLegacyResourceAccess" class="resource-manage-container">
          <div class="fixed-account-list-container">
            <AccountVendorGroup />
          </div>
          <RouterView class="router-view-content" />
        </div>
        <RouterView v-else />
      </template>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.main-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;

  .main-view {
    flex: 1;
    min-width: 1089px;
    min-height: 0;
    overflow-y: auto;

    &.status-view {
      display: flex;
      align-items: center;
      justify-content: center;
      min-width: auto;
      overflow: auto;
    }

    :deep(.page-container) {
      padding: 24px;
      height: 100%;
    }

    :deep(.detail-content-wrap) {
      padding: 24px;
      height: 100%;
      overflow-y: auto;

      .detail-info-main .info-list-item .item-field {
        width: 120px;
      }

      .detail-tab-main .info-title {
        margin-bottom: 8px;
        font-size: 14px;
      }
    }
  }
}
</style>
