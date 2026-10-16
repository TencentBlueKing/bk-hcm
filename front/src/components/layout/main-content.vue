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
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div :class="['main-view', { 'status-view': isStatusView }]">
      <RouterView :name="view" />
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
