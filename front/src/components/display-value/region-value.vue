<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useRegionStore } from '@/store/region';

const props = defineProps<{ value: string | string[]; vendor: string }>();

const list = ref([]);

const localValue = computed(() => {
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    return list.value.find((item) => item.id === id)?.name;
  });
  return names?.join?.(', ');
});

const regionStore = useRegionStore();

watchEffect(async () => {
  // TODO: 缓存与合并请求
  const res = await regionStore.getRegionList({ vendor: props.vendor });
  list.value = res;
});
</script>

<template>
  {{ displayValue }}
</template>
