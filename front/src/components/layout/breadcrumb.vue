<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useBack } from '@/router/hooks/use-back';
import type { RouteMetaConfig } from '@/router/meta';

const breadcrumb = useBreadcrumb();
const route = useRoute();
const { from, handleBack } = useBack();

const currentTitle = computed(() => {
  const routeMeta = route.meta as RouteMetaConfig;
  return breadcrumb.data.title ?? routeMeta?.menu?.i18n;
});

const showBack = computed(() => breadcrumb.data.back !== false && !!from.value);
</script>

<template>
  <div class="hcm-breadcrumb">
    <div class="breadcrumb-content">
      <i v-if="showBack" class="hcm-icon bkhcm-icon-arrows--left-line back-icon" @click="() => handleBack()" />
      <span class="breadcrumb-name">{{ currentTitle }}</span>
    </div>
    <div id="breadcrumbHead" class="breadcrumb-head"></div>
    <div id="breadcrumbExtra" class="breadcrumb-extra"></div>
  </div>
</template>

<style lang="scss" scoped>
.hcm-breadcrumb {
  display: flex;
  align-items: center;
  height: 52px;
  padding: 0 24px;
  background: #fff;
  box-shadow: 0 3px 4px 0 rgba(0, 0, 0, 0.04);
}

.breadcrumb-content {
  display: flex;
  align-items: center;

  .breadcrumb-name {
    font-size: 16px;
    letter-spacing: 0;
    color: #313238;
  }

  .back-icon {
    font-weight: bold;
    font-size: 16px;
    color: #3a84ff;
    cursor: pointer;
    margin-right: 10px;
  }
}

.breadcrumb-head {
  display: flex;
  align-items: center;
}

.breadcrumb-extra {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
</style>
