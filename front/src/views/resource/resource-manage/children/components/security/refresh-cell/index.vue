<template>
  <!-- loading -->
  <bk-loading v-if="loading" theme="primary" loading mode="spin" size="mini"></bk-loading>
  <div v-else class="refresh-cell">
    <!-- error -->
    <template v-if="showError">
      <i class="hcm-icon bkhcm-icon-alert error-icon" v-bk-tooltips="{ content: error, disabled: !showError }"></i>
      <bk-button class="text" theme="primary" text @click="emit('click')">{{ refreshText }}</bk-button>
    </template>
    <!-- 正常展示 -->
    <slot v-else></slot>
  </div>
</template>

<script setup lang="ts">
interface IProps {
  loading?: boolean;
  showError?: boolean;
  error?: string;
  refreshText?: string;
}

withDefaults(defineProps<IProps>(), {
  refreshText: '刷新',
});
const emit = defineEmits(['click']);
</script>

<style scoped lang="scss">
.refresh-cell {
  display: flex;
  align-items: center;

  .error-icon {
    color: $danger-color;
    font-size: 14px;
    cursor: pointer;
  }

  .text {
    margin-left: 4px;
    display: none;
  }

  &:hover {
    .text {
      display: block;
    }
  }
}
</style>
