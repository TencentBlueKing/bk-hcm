<script setup lang="ts">
import { useAttrs } from 'vue';
import type { ModelProperty, ModelPropertyType } from '@/model/typings';
import EnumValue from './enum-value.vue';
import StringValue from './string-value.vue';
import NumberValue from './number-value.vue';
import DatetimeValue from './datetime-value.vue';
import ArrayValue from './array-value.vue';
import BoolValue from './bool-value.vue';
import { DisplayType } from './typings';

defineOptions({ name: 'DisplayValue' });

const props = withDefaults(
  defineProps<{
    value: any;
    property: ModelProperty;
    display: DisplayType;
  }>(),
  {
    display: () => ({
      on: 'cell',
    }),
  },
);

const valueComps: Record<
  ModelPropertyType,
  typeof EnumValue | typeof StringValue | typeof DatetimeValue | typeof ArrayValue
> = {
  enum: EnumValue,
  datetime: DatetimeValue,
  number: NumberValue,
  string: StringValue,
  account: StringValue,
  user: StringValue,
  array: ArrayValue,
  bool: BoolValue,
};

const attrs = useAttrs();
</script>

<template>
  <component
    v-if="valueComps[property.type]"
    :is="valueComps[property.type]"
    :value="value"
    :option="property.option"
    :display="props.display"
    v-bind="attrs"
  />
  <span v-else>unknow type</span>
</template>
