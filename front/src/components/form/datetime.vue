<script setup lang="ts">
import { computed } from 'vue';
import { getDateShortcutRange } from '@/utils/search';
import type { DatePickerValueType } from 'bkui-vue/lib/date-picker/interface';

defineOptions({ name: 'hcm-form-datetime' });

const props = withDefaults(
  defineProps<{ format: string; type: 'date' | 'daterange' | 'datetime' | 'datetimerange' | 'month' | 'year' }>(),
  {
    format: 'yyyy-MM-dd HH:mm:ss',
  },
);

const rangeType = computed(() => ['daterange', 'datetimerange'].includes(props.type));
const shortcutsRange = computed(() => (rangeType.value ? getDateShortcutRange() : []));

const model = defineModel<DatePickerValueType>();

const localModel = computed({
  get: () => {
    if (!model.value) {
      return rangeType.value ? [] : ('' as unknown);
    }
    if (Array.isArray(model.value) && !model.value.filter((item) => Boolean(item)).length) {
      return [] as unknown;
    }
    // 当传入的值是dateString时，统一认为是ISO 8601的格式，透传给datepicker组件时未能正确展示为本地时间，转换为date时则正常
    if (typeof model.value === 'string') {
      return new Date(model.value) as DatePickerValueType;
    }
    if (Array.isArray(model.value)) {
      return model.value.map((item) => {
        if (typeof item === 'string') {
          return new Date(item);
        }
        return item;
      }) as DatePickerValueType;
    }
    return model.value;
  },
  set: (val: DatePickerValueType) => {
    model.value = val;
  },
});
</script>

<template>
  <bk-date-picker v-model="localModel" :type="type" :shortcuts="shortcutsRange" :format="format" />
</template>
