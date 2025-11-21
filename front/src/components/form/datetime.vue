<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import { getDateShortcutRange } from '@/utils/search';
import type { DatePickerValueType } from 'bkui-vue/lib/date-picker/interface';
import dayjs from 'dayjs';

interface IProps {
  format?: string;
  type?: 'date' | 'daterange' | 'datetime' | 'datetimerange' | 'month' | 'monthrange' | 'year';
  valueFormat?: string; // 返回值的格式 如：yyyy-MM-dd HH:mm:ss
  valueFormatter?: (val: any) => any;
}

interface IEmits {
  change: [val: DatePickerValueType | string | string[], originVal: any];
}

defineOptions({ name: 'hcm-form-datetime' });
const model = defineModel<DatePickerValueType | string | string[]>();
const props = withDefaults(defineProps<IProps>(), {
  format: 'yyyy-MM-dd HH:mm:ss',
  type: 'date',
});

const emits = defineEmits<IEmits>();

const rangeType = computed(() => ['daterange', 'datetimerange'].includes(props.type));
const shortcutsRange = computed(() => (rangeType.value ? getDateShortcutRange(props.type !== 'daterange') : []));

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
    model.value = formatValue(val);
  },
});

const attrs = useAttrs();

// 用于外部获取格式化后的值
const formatValue = (val: DatePickerValueType | string | string[]) => {
  if (props.valueFormat) {
    if (Array.isArray(val)) {
      return val.map((item) => dayjs(item).format(props.valueFormat));
    }
    return dayjs(val).format(props.valueFormat);
  }
  if (props.valueFormatter) {
    return props.valueFormatter(val);
  }
  return val;
};

const handleChange = (val: DatePickerValueType | string | string[]) => {
  emits('change', formatValue(val), val);
};
</script>

<template>
  <bk-date-picker
    v-model="localModel"
    v-bind="attrs"
    :type="type"
    :shortcuts="shortcutsRange"
    :format="format"
    @change="handleChange"
  />
</template>
