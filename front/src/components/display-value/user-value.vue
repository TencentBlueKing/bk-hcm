<script setup lang="ts">
defineOptions({ name: 'hcm-user-value' });
defineProps<{ value: string | string[] }>();
</script>

<template>
  <bk-user-display-name :user-id="value" />
</template>
