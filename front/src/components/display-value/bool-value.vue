<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';

const props = defineProps<{ value: boolean; option?: ModelProperty['option'] }>();

const trueText = computed(() => props.option?.trueText ?? props.value);
const falseText = computed(() => props.option?.falseText ?? props.value);
const displayValue = computed(() => (props.value ? trueText.value : falseText.value));
</script>

<template>
  {{ displayValue ?? '--' }}
</template>
