<script setup lang="ts">
import { computed } from 'vue';
import { AppearanceType, DisplayType } from './typings';
import { useBusinessGlobalStore } from '@/store/business-global';

import BusinessAssignTag from './appearance/business-assign-tag.vue';

const props = defineProps<{ value: number | number[]; separator?: string; display?: DisplayType }>();

const businessGlobalStore = useBusinessGlobalStore();

const appearance = computed(() => props.display?.appearance);

const appearanceComps: Partial<Record<AppearanceType, any>> = {
  'business-assign-tag': BusinessAssignTag,
};

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
  <template v-if="!appearance">
    <bk-overflow-title resizeable type="tips" v-if="display?.showOverflowTooltip">{{ displayValue }}</bk-overflow-title>
    <span v-else>{{ displayValue }}</span>
  </template>
  <component v-else :is="appearanceComps[appearance]" :display-value="displayValue" :value="value" />
</template>
