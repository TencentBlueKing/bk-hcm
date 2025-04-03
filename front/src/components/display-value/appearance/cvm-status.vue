<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { DisplayType } from '../typings';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '@/views/resource/resource-manage/common/table/HostOperations';

const props = defineProps<{
  value: string;
  displayValue: string | number | string[] | number[];
  option: ModelProperty['option'];
  displayOn: DisplayType['on'];
}>();

const icon = computed(() => {
  const status = props.value;
  if (HOST_RUNNING_STATUS.includes(status)) return StatusNormal;
  if (HOST_SHUTDOWN_STATUS.includes(status)) return status === 'stopped' ? StatusUnknown : StatusAbnormal;
  return StatusUnknown;
});
</script>

<template>
  <div class="cvm-status">
    <img :src="icon" class="icon" alt="icon" />
    <bk-overflow-title resizeable type="tips">
      <span class="text">{{ displayValue }}</span>
    </bk-overflow-title>
  </div>
</template>

<style lang="scss" scoped>
.cvm-status {
  display: flex;
  align-items: center;

  .icon {
    width: 14px;
    height: 14px;
    margin-right: 4px;
  }
}
</style>
