<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import HcmBreadcrumb from './breadcrumb.vue';

const route = useRoute();
const breadcrumb = useBreadcrumb();

const showBreadcrumb = computed(() => breadcrumb.data.display);

const view = computed(() => route.meta.view);
</script>

<template>
  <div class="main-content">
    <HcmBreadcrumb v-if="showBreadcrumb" />
    <div class="main-layout" :class="{ 'no-breadcrumb': !showBreadcrumb }">
      <RouterView v-slot="{ Component }" class="main-view" :name="view">
        <component :is="Component" />
      </RouterView>
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
    overflow: auto;

    .main-view {
      min-width: 1089px;
      height: 100%;
      overflow: hidden;
    }

    &.no-breadcrumb {
      height: 100%;
    }

    :deep(.detail-content-wrap) {
      padding: 24px;
      height: 100%;

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
