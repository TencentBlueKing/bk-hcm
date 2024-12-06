<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusLoading from '@/assets/image/status_loading.png';
import StatusFailure from '@/assets/image/failed-account.png';
import { DisplayType } from '../typings';

const props = defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  option: ModelProperty['option'];
  displayOn: DisplayType['on'];
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
    case 'running':
      return StatusLoading;
    case 'abnormal':
      return StatusAbnormal;
    case 'normal':
      return StatusNormal;
    default:
      return StatusUnknown;
  }
});
</script>

<template>
  <div class="status">
    <img :src="icon" :class="['icon', props.value]" alt="icon" />
    <bk-overflow-title resizeable type="tips">
      <span class="text">{{ displayValue }}</span>
    </bk-overflow-title>
  </div>
</template>

<style lang="scss" scoped>
.status {
  display: flex;
  align-items: center;

  .icon {
    width: 14px;
    height: 14px;
    margin-right: 4px;

    &.running {
      width: 12px;
      height: 12px;
    }
  }
}
</style>
