<script setup lang="ts">
import { watchEffect } from 'vue';

defineOptions({ name: 'hcm-search-datetime' });

const props = withDefaults(defineProps<{ shortcutSelectedIndex: number; format: string; useShortcutText: boolean }>(), {
  shortcutSelectedIndex: 1,
  format: 'yyyy-MM-dd HH:mm:ss',
  useShortcutText: true,
});

const shortcutsRange = [
  {
    text: '今天',
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: '近7天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: '近15天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: '近30天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  },
];

const model = defineModel<Date[]>();

watchEffect(() => {
  if (props.shortcutSelectedIndex >= 0 && props.shortcutSelectedIndex < shortcutsRange.length) {
    model.value = shortcutsRange[props.shortcutSelectedIndex].value();
  }
});
</script>

<template>
  <bk-date-picker
    v-model="model"
    type="datetimerange"
    :shortcut-selected-index="shortcutSelectedIndex"
    :shortcuts="shortcutsRange"
    :format="format"
    :use-shortcut-text="useShortcutText"
  />
</template>
