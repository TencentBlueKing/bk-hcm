<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { AppearanceType, DisplayType } from './typings';
import Status from './appearance/status.vue';

const props = defineProps<{
  value: string | number | string[] | number[];
  option: ModelProperty['option'];
  display: DisplayType;
}>();

const displayOn = computed(() => props.display?.on || 'cell');
const appearance = computed(() => props.display?.appearance);

const displayValue = computed(() => {
  const vals = Array.isArray(props.value) ? props.value : [props.value];
  return vals.map((val) => props.option?.[val] || props.value).join(', ') || '--';
});

const appearanceComps: Record<AppearanceType, any> = {
  status: Status,
};
</script>

<template>
  <component
    :is="appearanceComps[appearance]"
    v-if="appearance"
    :display-value="displayValue"
    :display-on="displayOn"
    :value="value"
    :option="option"
  />
  <span v-else>{{ displayValue }}</span>
</template>
