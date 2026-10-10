<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import AccountVendorGroup from '@/views/resource/resource-manage/account/vendor-group/index.vue';
import HcmBreadcrumb from './breadcrumb.vue';

defineProps<{ isRouterAlive: boolean }>();

const route = useRoute();
const breadcrumb = useBreadcrumb();

// 守卫按权限 / 异常把 meta.view 切到 permission、error、notFound，对应 views/index.ts 注入的 named view
const view = computed(() => (route.meta.view as string) ?? 'default');
// TODO(菜单整合二期·组装)：原资源接入页面全部迁出后删除
const isLegacyResourceAccess = computed(() => !!route.meta.legacyResourceAccess);
const isStatusView = computed(() => view.value !== 'default' || route.name === '404' || route.name === 'error');
const showBreadcrumb = computed(() => !isLegacyResourceAccess.value && breadcrumb.data.display);
// TODO(菜单整合二期·组装)：原资源详情迁出后删除
// 原资源详情页在右侧内容区显示面包屑；返回列表时带上资源类型以回到对应页签
const legacyBreadcrumbRoutes = ['resourceDetail'];
const showLegacyBreadcrumb = computed(() => legacyBreadcrumbRoutes.includes(route.name));
// 原详情类型与资源纳管页签不同名的映射
const legacyDetailTabMap: Record<string, string> = {
  eips: 'ip',
  route: 'routing',
  gcp: 'security',
  template: 'security',
};
const legacyBackConfig = computed(() => {
  if (route.name !== 'resourceDetail') return undefined;
  const type = route.params.type as string;
  return { params: { resourceType: legacyDetailTabMap[type] ?? type }, query: { accountId: route.query.accountId } };
});
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div :class="['main-view', 'view-warp', { 'status-view': isStatusView }]">
      <template v-if="isRouterAlive">
        <!-- 状态页不展示云账号树，避免无权限时露出页面数据 -->
        <div v-if="isLegacyResourceAccess && view === 'default'" class="resource-manage-container">
          <div class="fixed-account-list-container">
            <AccountVendorGroup />
          </div>
          <div v-if="showLegacyBreadcrumb" class="router-view-content legacy-breadcrumb-view">
            <HcmBreadcrumb :back-config="legacyBackConfig" />
            <RouterView class="legacy-breadcrumb-page" />
          </div>
          <RouterView v-else class="router-view-content" />
        </div>
        <RouterView v-else :name="view" />
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
