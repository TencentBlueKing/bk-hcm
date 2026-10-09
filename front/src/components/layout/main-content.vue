<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import {
  MENU_RESOURCE,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_OPERATION_LOG_DETAILS,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
} from '@/constants/menu-symbol';
import AccountVendorGroup from '@/views/resource/resource-manage/account/vendor-group/index.vue';
import HcmBreadcrumb from './breadcrumb.vue';

defineProps<{ isRouterAlive: boolean }>();

const route = useRoute();
const breadcrumb = useBreadcrumb();

// TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除
const isLegacyResourceAccess = computed(() => route.matched[0]?.name === MENU_RESOURCE);
const isStatusView = computed(() => route.name === '404' || route.name === 'error');
const showBreadcrumb = computed(() => !isLegacyResourceAccess.value && breadcrumb.data.display);
// TODO(菜单整合二期·组装)：资源接入迁入资源运营后删除，由资源纳管入口右栏的面包屑承担
// 资源接入的详情 / 申请页在右侧内容区显示面包屑；详情返回列表时带上资源类型以回到对应页签
const legacyBreadcrumbRoutes = [
  'resourceDetail',
  MENU_RESOURCE_OPERATION_LOG_DETAILS,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_VPC_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
];
const showLegacyBreadcrumb = computed(() => legacyBreadcrumbRoutes.includes(route.name));
const legacyBackConfig = computed(() =>
  route.name === 'resourceDetail' ? { query: { type: route.params.type as string } } : undefined,
);
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div :class="['main-view', 'view-warp', { 'status-view': isStatusView }]">
      <template v-if="isRouterAlive">
        <div v-if="isLegacyResourceAccess" class="resource-manage-container">
          <div class="fixed-account-list-container">
            <AccountVendorGroup />
          </div>
          <div v-if="showLegacyBreadcrumb" class="router-view-content legacy-breadcrumb-view">
            <HcmBreadcrumb :back-config="legacyBackConfig" />
            <RouterView class="legacy-breadcrumb-page" />
          </div>
          <RouterView v-else class="router-view-content" />
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

    // TODO(菜单整合二期·组装)：随资源接入过渡布局删除
    // 资源接入右侧内容区：面包屑固定在顶部，页面自带内边距并独立滚动
    .resource-manage-container .legacy-breadcrumb-view {
      display: flex;
      flex-direction: column;
      padding: 0;
      overflow: hidden;

      .legacy-breadcrumb-page {
        flex: 1;
        min-height: 0;
        overflow-y: auto;
      }
    }
  }
}
</style>
