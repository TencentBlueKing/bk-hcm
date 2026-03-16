<script setup lang="ts">
import { inject } from 'vue';

const _isBusinessPage = inject('isBusinessPage', false);
const _isServicePage = inject('isServicePage', false);
</script>

<template>
  <div class="gpu-demand-detail">GPU需求详情 - 开发中</div>
</template>

<style lang="scss" scoped>
.gpu-demand-detail {
  height: 100%;
}
</style>
