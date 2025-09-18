<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';
import { AppearanceType, DisplayType } from './typings';
import Status from './appearance/status.vue';
import CvmStatus from './appearance/cvm-status.vue';
import ClbStatus from './appearance/clb-status.vue';
import DynamicStatus from './appearance/dynamic-status.vue';

const props = defineProps<{
  value: string | number | string[] | number[];
  option: ModelProperty['option'];
  display: DisplayType;
}>();

const localOption = ref<Record<string | number, any>>({});

watchEffect(async () => {
  if (typeof props.option === 'function') {
    localOption.value = await props.option();
  } else {
    localOption.value = props.option;
  }
});

const displayOn = computed(() => props.display?.on || 'cell');
const appearance = computed(() => props.display?.appearance);
const appearanceProps = computed(() => props.display?.appearanceProps);

const displayValue = computed(() => {
  const vals = Array.isArray(props.value) ? props.value : [props.value];
  return vals.map((val) => localOption.value?.[val] || val).join(', ') || '--';
});

const appearanceComps: Partial<Record<AppearanceType, any>> = {
  status: Status,
  'cvm-status': CvmStatus,
  'clb-status': ClbStatus,
  'dynamic-status': DynamicStatus,
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
    v-bind="appearanceProps"
  />
  <span v-else>{{ displayValue }}</span>
</template>
