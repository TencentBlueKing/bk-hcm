<script setup lang="ts">
import { computed } from 'vue';
import { DisplayType } from './typings';
import { useBusinessGlobalStore } from '@/store/business-global';

const businessGlobalStore = useBusinessGlobalStore();

const props = defineProps<{ value: number | number[]; separator?: string; display?: DisplayType }>();

const displayValue = computed(() => {
  const values = Array.isArray(props.value) ? props.value : [props.value];
  const names = [];
  for (const value of values) {
    const name = businessGlobalStore.businessFullList.find((item) => item.id === value)?.name;
    names.push(name);
  }
  return names?.join?.(props.separator || ', ') || '--';
});
</script>

<template>
  <bk-overflow-title resizeable type="tips" v-if="display?.showOverflowTooltip">{{ displayValue }}</bk-overflow-title>
  <span v-else>{{ displayValue }}</span>
</template>
