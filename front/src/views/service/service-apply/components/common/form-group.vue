<script setup lang="ts">
defineProps<{
  title: string;
}>();
</script>
<template>
  <div class="form-group">
    <div class="group-title" v-if="title">{{ title }}</div>
    <div class="group-content">
      <slot></slot>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.form-group {
  .group-title {
    font-size: 14px;
    font-weight: normal;
    padding: 8px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.15);
  }

  .group-content {
    padding: 12px 36px 0 0;
    width: 60%;
  }

  & + .form-group {
    margin-top: 4px;
  }
}

@media screen and (max-width: 1660px) {
  .form-group {
    .group-content {
      width: 100%;
    }
  }
}
</style>
