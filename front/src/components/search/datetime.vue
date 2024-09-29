<script setup lang="ts">
import { getDateShortcutRange } from '@/utils/search';

defineOptions({ name: 'hcm-search-datetime' });

withDefaults(defineProps<{ format: string }>(), {
  format: 'yyyy-MM-dd HH:mm:ss',
});

const shortcutsRange = getDateShortcutRange();

const model = defineModel<Date[] | string[]>();
</script>

<template>
  <bk-date-picker v-model="model" type="datetimerange" :shortcuts="shortcutsRange" :format="format" />
</template>
