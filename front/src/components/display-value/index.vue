<script setup lang="ts">
import { useAttrs } from 'vue';
import type { ModelProperty, ModelPropertyType } from '@/model/typings';
import EnumValue from './enum-value.vue';
import StringValue from './string-value.vue';
import NumberValue from './number-value.vue';

defineOptions({ name: 'DisplayValue' });

defineProps<{ value: any; property: ModelProperty }>();

const valueComps: Record<ModelPropertyType, typeof EnumValue | typeof StringValue> = {
  enum: EnumValue,
  datetime: StringValue,
  number: NumberValue,
  string: StringValue,
  account: StringValue,
  user: StringValue,
};

const attrs = useAttrs();
</script>

<template>
  <component :is="valueComps[property.type]" :value="value" :option="property.option" v-bind="attrs" />
</template>
