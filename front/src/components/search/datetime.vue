<script setup lang="ts">
import DateTime from '@/components/form/datetime.vue';
import type { DatePickerValueType } from 'bkui-vue/lib/date-picker/interface';

defineOptions({ name: 'hcm-search-datetime' });

const model = defineModel<DatePickerValueType>();

withDefaults(
  defineProps<{
    format: string;
    type: 'date' | 'daterange' | 'datetime' | 'datetimerange' | 'month' | 'year';
    appendToBody: boolean;
  }>(),
  {
    format: 'yyyy-MM-dd HH:mm:ss',
    type: 'datetimerange',
    appendToBody: true,
  },
);
</script>

<template>
  <DateTime v-model="model" :type="type" :format="format" :append-to-body="appendToBody" />
</template>
