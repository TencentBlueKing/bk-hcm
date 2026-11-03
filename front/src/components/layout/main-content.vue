<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import HcmBreadcrumb from './breadcrumb.vue';

const route = useRoute();
const breadcrumb = useBreadcrumb();

// 守卫按权限 / 异常把 meta.view 切到 permission、error、notFound，对应 views/index.ts 注入的 named view
const view = computed(() => (route.meta.view as string) ?? 'default');
const isStatusView = computed(() => view.value !== 'default' || route.name === '404' || route.name === 'error');
const showBreadcrumb = computed(() => breadcrumb.data.display);
const isFitWidth = computed(() => route.meta.layout?.contentWidth === 'fit');
</script>

<template>
  <div class="main-content">
    <div class="main-scroll" :class="{ 'is-fit-width': isFitWidth }">
      <div class="main-layout" :class="{ 'is-status-view': isStatusView, 'is-fit-width': isFitWidth }">
        <HcmBreadcrumb v-if="showBreadcrumb" />
        <div :class="['main-view', { 'status-view': isStatusView, 'is-fit-width': isFitWidth }]">
          <RouterView :name="view" />
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.main-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;

  .main-scroll {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: auto hidden;

    &.is-fit-width {
      overflow: hidden;
    }
  }

  .main-layout {
    display: flex;
    flex-direction: column;

    // 与 html/body 的 980px 兜底取同一数值；注意主区在侧栏（260px）之后，
    // 所以非 fit 页面的整页下限约 1240px，业务视角工具条未适配的问题见 app-shell 的遗留项
    min-width: 980px;
    height: 100%;

    &.is-status-view,
    &.is-fit-width {
      min-width: 0;
    }

    :deep(.hcm-breadcrumb) {
      flex-shrink: 0;
    }
  }

  .main-view {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: auto;

    &.is-fit-width {
      overflow: hidden auto;
    }

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
