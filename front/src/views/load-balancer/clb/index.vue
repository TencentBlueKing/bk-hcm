<script setup lang="ts">
import { useTemplateRef } from 'vue';
import LoadBalancerList from './load-balancer-list.vue';

defineOptions({ name: 'load-balancer-view' });

// TODO-CLB：这里存在一个定位问题（url访问详情页没法定位。可能是virtual-render刚挂载的时候，它的高度计算有问题，第一次的滚动会失效）
const loadBalancerListRef = useTemplateRef<typeof LoadBalancerList>('load-balancer-list');
const handleDetailsShow = (id: string) => {
  loadBalancerListRef.value?.fixToActive(id);
};
</script>

<template>
  <bk-resize-layout class="container" collapsible :initial-divide="300" :min="300">
    <template #aside>
      <load-balancer-list ref="load-balancer-list" />
    </template>
    <template #main>
      <!-- load-balancer-overview -->
      <!-- load-balancer-details -->
      <router-view @details-show="handleDetailsShow" />
    </template>
  </bk-resize-layout>
</template>

<style scoped lang="scss">
.container {
  height: 100%;
}
</style>
