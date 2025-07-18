<script setup lang="ts">
import TargetGroupList from '@/views/business/load-balancer/group-view/target-group-list';

defineOptions({ name: 'load-balancer-view' });
</script>

<template>
  <bk-resize-layout class="container" collapsible :initial-divide="300" :min="300">
    <template #aside>
      <target-group-list ref="load-balancer-list" />
    </template>
    <template #main>
      <!-- target-group-overview -->
      <!-- target-group-details -->
      <router-view />
    </template>
  </bk-resize-layout>
</template>

<style scoped lang="scss">
.container {
  height: 100%;
}
</style>
