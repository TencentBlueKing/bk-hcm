<script setup lang="ts">
import { ref } from 'vue';

import ActionDropdown from './action-dropdown/index.vue';
import OperationSideslider from './operation-sideslider/index.vue';

import { Action } from './types';

// 下拉菜单
const activeAction = ref<Action>();

// 抽屉
const isSliderShow = ref(false);
const showSideslider = (action: Action) => {
  activeAction.value = action;
  isSliderShow.value = true;
};

defineOptions({ name: 'LbBatchImportComp' });
</script>

<template>
  <!-- 下拉菜单 -->
  <ActionDropdown @click="showSideslider" />

  <!-- 抽屉 -->
  <OperationSideslider v-model="isSliderShow" :active-action="activeAction" />
</template>

<style scoped lang="scss"></style>
