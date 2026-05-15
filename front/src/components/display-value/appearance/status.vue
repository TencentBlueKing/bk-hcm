<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { Spinner } from 'bkui-vue/lib/icon';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import { DisplayType } from '../typings';

const props = defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  option?: ModelProperty['option'];
  displayOn?: DisplayType['on'];
}>();

const icon = computed(() => {
  switch (props.value) {
    case 'success':
      return StatusSuccess;
    case 'failure':
    case 'failed':
    case 'fail':
    case 'deliver_partial':
      return StatusFailure;
    case 'abnormal':
      return StatusAbnormal;
    case 'normal':
    case 'enabled':
      return StatusNormal;
    case 'disabled':
    default:
      return StatusUnknown;
  }
});

const isLoading = computed(() => {
  return props.value === 'running';
});
</script>

<template>
  <div class="status">
    <Spinner v-if="isLoading" fill="#3A84FF" :width="14" :height="14" class="status-icon" />
    <img v-else :src="icon" class="icon" alt="icon" />
    <bk-overflow-title resizeable type="tips">
      <span class="text">{{ displayValue }}</span>
    </bk-overflow-title>
  </div>
</template>

<style lang="scss" scoped>
.status {
  display: flex;
  align-items: center;

  .status-icon {
    margin-right: 4px;
    flex-shrink: 0;
  }

  .icon {
    width: 14px;
    height: 14px;
    margin-right: 4px;
    flex-shrink: 0;
  }
}
</style>
