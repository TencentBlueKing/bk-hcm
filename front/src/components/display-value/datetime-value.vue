<script setup lang="ts">
import { computed } from 'vue';
import { timeFormatter } from '@/common/util';
const props = withDefaults(defineProps<{ value: string; format: string }>(), {
  format: 'YYYY-MM-DD HH:mm:ss',
});

const displayValue = computed(() => timeFormatter(props.value, props.format));
</script>

<template>
  {{ displayValue || '--' }}
</template>
