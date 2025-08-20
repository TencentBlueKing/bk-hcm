<script setup lang="ts">
const props = defineProps<{ step: number; title: string }>();
</script>

<template>
  <div class="step-container">
    <div class="title">
      <span class="step-num">Step {{ props.step }}</span>
      <span>{{ props.title }}</span>
    </div>
    <div class="content">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped lang="scss">
.step-container {
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

  & + .step-container {
    margin-top: 32px;
  }
}
</style>
