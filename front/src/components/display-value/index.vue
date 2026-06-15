<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import type { ModelProperty, ModelPropertyType, PropertyDisplayConfig } from '@/model/typings';
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

defineOptions({ name: 'DisplayValue' });

const props = withDefaults(
  defineProps<{
    value: any;
    property: ModelProperty;
    display?: PropertyDisplayConfig;
  }>(),
  {
    display: () => ({
      on: 'cell',
    }),
  },
);

// 获取自定义 render 函数，优先从 display props 获取，其次从 property.meta.display 获取
const customRender = computed(() => props.display?.render || props.property.meta?.display?.render);

// 计算自定义渲染结果
const customRenderResult = computed(() => (customRender.value ? customRender.value(props.value) : null));

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
  <!-- 优先使用自定义 render -->
  <component v-if="customRender" :is="() => customRenderResult" />
  <!-- 否则使用类型对应的组件 -->
  <component
    v-else-if="valueComps[property.type]"
    :is="valueComps[property.type]"
    :value="value"
    :option="property.option"
    :display="props.display"
    v-bind="attrs"
  >
    <template v-for="(_, slot) of $slots" #[slot]="scope">
      <slot :name="slot" v-bind="scope" />
    </template>
  </component>
  <span v-else>unknown type</span>
</template>
