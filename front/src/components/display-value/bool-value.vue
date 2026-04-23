<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';

const props = defineProps<{ value: boolean; option?: ModelProperty['option'] }>();

const localOption = ref<Record<string | number, any>>({});

watchEffect(async () => {
  if (typeof props.option === 'function') {
    localOption.value = await props.option();
  } else {
    localOption.value = props.option;
  }
});

const trueText = computed(() => localOption.value?.trueText ?? props.value);
const falseText = computed(() => localOption.value?.falseText ?? props.value);
const displayValue = computed(() => (props.value ? trueText.value : falseText.value));
</script>

<template>
  {{ displayValue ?? '--' }}
</template>
