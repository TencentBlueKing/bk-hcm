<template>
  <bk-alert>
    <template #title>
      <p>所选机型为{{ info.device_type }}，CPU为{{ info.cpu_core }}核，内存为{{ info.memory }}G。</p>
      <slot></slot>
      <p v-if="helperText">{{ helperText }}</p>
    </template>
  </bk-alert>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { CvmDeviceType } from './types';

interface IProps {
  info: CvmDeviceType;
  isDefaultFourYears?: boolean;
  isGpuDeviceType?: boolean;
}

defineOptions({ name: 'cvm-device-type-tip' });
const props = defineProps<IProps>();

const helperText = computed(() => {
  const { isDefaultFourYears, isGpuDeviceType } = props;
  if (isGpuDeviceType) {
    return 'GPU类机型只能选择6年套餐，按100%折扣比例计费（预测外需先选择”按量计费“后再转为”包年包月“6年套餐）';
  }
  if (isDefaultFourYears) {
    return '专用机型只能选择4年套餐，按100%折扣比例计费（预测外需先选择”按量计费“后再转为”包年包月“4年套餐）';
  }
  return '';
});
</script>
