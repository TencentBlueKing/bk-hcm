<script setup lang="ts">
import { provide } from 'vue';
import LoadBalancerManage from '../resource/resource-manage/children/manage/load-balancer-manage.vue';

provide('currentGlobalBusinessId', {});
</script>

<template>
  <load-balancer-manage>
    <slot></slot>
  </load-balancer-manage>
</template>

<style scoped></style>
