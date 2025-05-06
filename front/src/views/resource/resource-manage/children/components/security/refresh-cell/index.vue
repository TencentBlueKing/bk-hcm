<template>
  <div class="refresh-cell">
    <bk-loading v-if="loading" theme="primary" loading mode="spin" size="mini"></bk-loading>
    <template v-else>
      <i class="hcm-icon bkhcm-icon-alert error-icon" v-bk-tooltips="{ content: error, disabled: !error }"></i>
      <bk-button class="text" theme="primary" text @click="emit('click')">{{ text }}</bk-button>
    </template>
  </div>
</template>

<script setup lang="ts">
interface IProps {
  loading?: boolean;
  error?: string;
  text?: string;
}

withDefaults(defineProps<IProps>(), {
  text: '刷新',
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
