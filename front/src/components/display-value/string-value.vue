<script setup lang="ts">
import { computed } from 'vue';
import { AppearanceType, DisplayType } from './typings';
import Link from './appearance/link.vue';
import { isNil, isString } from 'lodash';

const props = defineProps<{ value: string | number | string[] | number[]; display: DisplayType }>();

const displayOn = computed(() => props.display?.on || 'cell');
const appearance = computed(() => props.display?.appearance);
const format = computed(() => props.display?.format);

const displayValue = computed(() => {
  if (isNil(props.value) || (isString(props.value) && !props.value)) return '--';

  const vals = Array.isArray(props.value) ? props.value : [props.value];

  if (typeof format.value === 'function') {
    const formattedVals = vals.map((item) => format.value(item));
    return formattedVals.join(', ');
  }

  return vals.join(', ');
});

const appearanceComps: Partial<Record<AppearanceType, any>> = {
  link: Link,
};
</script>

<template>
  <template v-if="!appearance">
    <bk-overflow-title class="full-width" resizeable type="tips" v-if="display?.showOverflowTooltip">
      {{ displayValue }}
    </bk-overflow-title>
    <span v-else>{{ displayValue }}</span>
  </template>
  <component
    v-else
    :is="appearanceComps[appearance]"
    :display-value="displayValue"
    :display-on="displayOn"
    :value="value"
  />
</template>
