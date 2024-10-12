<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ value: string[] }>();

const displayValue = computed(() => {
  const vals = Array.isArray(props.value) ? props.value : [props.value];
  return vals?.join?.(', ') || '--';
});
</script>

<template>
  {{ displayValue }}
</template>
