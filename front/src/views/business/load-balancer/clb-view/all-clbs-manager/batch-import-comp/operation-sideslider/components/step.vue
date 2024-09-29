<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  step: number;
  title: string;
}>();

const { t } = useI18n();
</script>

<template>
  <div class="step-wrapper">
    <div class="title">
      <span class="step-num">{{ t(`Step ${props.step}`) }}</span>
      <span>{{ props.title }}</span>
    </div>
    <div class="content">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped lang="scss">
.step-wrapper {
  .title {
    margin-bottom: 12px;
    font-size: 14px;
    color: #313238;

    .step-num {
      margin-right: 12px;
      color: #ea3636;
    }
  }

  .content {
    margin-bottom: 30px;
  }

  & + .step-wrapper {
    margin-top: 32px;
  }
}
</style>
