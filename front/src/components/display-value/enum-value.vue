<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';

const props = defineProps<{ value: string | number | string[] | number[]; option: ModelProperty['option'] }>();

const displayValue = computed(() => {
  const vals = Array.isArray(props.value) ? props.value : [props.value];
  return vals.map((val) => props.option?.[val] || props.value).join(', ') || '--';
});
</script>

<template>
  {{ displayValue }}
</template>
