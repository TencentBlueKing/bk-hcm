<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import StatusLoading from '@/assets/image/status_loading.png';
import { DisplayType } from '../typings';

const props = defineProps<{
  value: string | number;
  displayValue: string | number;
  option: ModelProperty['option'];
  displayOn: DisplayType['on'];
  statusObject: Record<'success' | 'fail' | 'wait' | 'ing' | 'stop', Array<string | number>>;
}>();

const status = computed(() => {
  let status = 'unknown';
  for (const [key, value] of Object.entries(props.statusObject)) {
    if (value.includes(props.value)) {
      status = key;
    }
  }
  return status;
});
</script>

<template>
  <div class="status">
    <span :class="['icon', status]" v-if="status !== 'ing'"></span>
    <img :src="StatusLoading" :class="['icon', status]" alt="icon" v-else />
    <bk-overflow-title resizeable type="tips">
      <span class="text">{{ displayValue }}</span>
    </bk-overflow-title>
  </div>
</template>

<style lang="scss" scoped>
.status {
  display: flex;
  align-items: center;
  gap: 10px;

  .icon {
    width: 8px;
    height: 8px;
    border-radius: 50%;

    &.success {
      background: #cbf0da;
      border: 1px solid #2caf5e;
    }

    &.fail {
      background: #fdd;
      border: 1px solid #ea3636;
    }

    &.wait {
      background: #fce5c0;
      border: 1px solid #f59500;
    }

    &.stop {
      background: #f0f1f5;
      border: 1px solid #c4c6cc;
    }

    &.ing {
      width: 10px;
      height: 10px;
      animation: spin 2s linear infinite;
    }
  }
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}
</style>
