<script setup lang="ts">
import { computed } from 'vue';
import { isNil } from 'lodash';
import type { DisplayType } from './typings';

const props = withDefaults(defineProps<{ value: string | number; display?: DisplayType }>(), {
  display: () => ({}),
});

const displayValue = computed(() => {
  if (isNil(props.value) || props.value === '') return '--';
  const { format } = props.display || {};
  return typeof format === 'function' ? format(props.value) : props.value;
});
</script>

<template>
  {{ displayValue }}
</template>
