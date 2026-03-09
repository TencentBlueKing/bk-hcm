<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import HcmBreadcrumb from './breadcrumb.vue';

const route = useRoute();
const breadcrumb = useBreadcrumb();

const view = computed(() => route.meta.view);
const isStatusPage = computed(() => route.name === '404' || route.name === 'error');
const isStatusView = computed(() => isStatusPage.value || (view.value && view.value !== 'default'));
const showBreadcrumb = computed(() => breadcrumb.data.display);
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div class="main-layout" :class="{ 'no-breadcrumb': !showBreadcrumb }">
      <div class="main-view" :class="{ 'status-view': isStatusView }">
        <RouterView :name="view" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.main-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;

  .main-layout {
    height: calc(100% - 52px);
    overflow: hidden;

    .main-view {
      min-width: 1089px;
      height: 100%;
      overflow: hidden;

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

    &.no-breadcrumb {
      height: 100%;
    }
  }
}
</style>
