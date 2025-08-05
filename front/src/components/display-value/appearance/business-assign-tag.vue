<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  value: number | number[];
  displayValue: string;
}>();

const hasBusiness = computed(() => {
  if (Array.isArray(props.value) && props.value.length > 0) {
    return true;
  }
  if (!Array.isArray(props.value) && props.value) {
    return true;
  }
  return false;
});

const theme = computed(() => (hasBusiness.value ? 'success' : 'default'));

const tooltips = computed(() => {
  if (hasBusiness.value) {
    return { content: props.displayValue };
  }
  return { disabled: true };
});
</script>

<template>
  <bk-tag :theme="theme" v-bk-tooltips="tooltips">{{ hasBusiness ? displayValue : '未分配' }}</bk-tag>
</template>
