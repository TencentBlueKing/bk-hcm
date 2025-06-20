<script setup lang="ts">
import { useAttrs } from 'vue';
import type { ModelProperty, ModelPropertyType } from '@/model/typings';
import EnumValue from './enum-value.vue';
import StringValue from './string-value.vue';
import NumberValue from './number-value.vue';
import DatetimeValue from './datetime-value.vue';
import ArrayValue from './array-value.vue';
import BoolValue from './bool-value.vue';
import CertValue from './cert-value.vue';
import CaValue from './ca-value.vue';
import RegionValue from './region-value.vue';
import BusinessValue from './business-value.vue';
import UserValue from './user-value.vue';
import CloudAreaValue from './cloud-area-value.vue';
import JsonValue from './json-value.vue';
import { DisplayType } from './typings';

defineOptions({ name: 'DisplayValue' });

const props = withDefaults(
  defineProps<{
    value: any;
    property: ModelProperty;
    display?: DisplayType;
  }>(),
  {
    display: () => ({
      on: 'cell',
    }),
  },
);

const valueComps: Record<
  ModelPropertyType,
  | typeof EnumValue
  | typeof DatetimeValue
  | typeof NumberValue
  | typeof StringValue
  | typeof ArrayValue
  | typeof BoolValue
  | typeof CertValue
  | typeof CaValue
  | typeof RegionValue
  | typeof BusinessValue
  | typeof UserValue
  | typeof CloudAreaValue
  | typeof JsonValue
> = {
  enum: EnumValue,
  datetime: DatetimeValue,
  number: NumberValue,
  string: StringValue,
  account: StringValue,
  array: ArrayValue,
  bool: BoolValue,
  cert: CertValue,
  ca: CaValue,
  region: RegionValue,
  business: BusinessValue,
  json: JsonValue,
  user: UserValue,
  'cloud-area': CloudAreaValue,
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
  <span v-else>unknown type</span>
</template>
