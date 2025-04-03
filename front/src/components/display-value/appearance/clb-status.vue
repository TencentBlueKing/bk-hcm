<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { DisplayType } from '../typings';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusLoading from '@/assets/image/status_loading.png';

const props = defineProps<{
  value: string;
  displayValue: string | number | string[] | number[];
  option: ModelProperty['option'];
  displayOn: DisplayType['on'];
}>();

const icon = computed(() => {
  const status = props.value;
  if (status === '0') return StatusLoading;
  return StatusSuccess;
});
</script>

<template>
  <div class="clb-status">
    <img :src="icon" class="icon" :class="{ 'spin-icon': value === '0' }" alt="icon" />
    <bk-overflow-title resizeable type="tips">
      <span class="text">{{ displayValue }}</span>
    </bk-overflow-title>
  </div>
</template>

<style lang="scss" scoped>
.clb-status {
  display: flex;
  align-items: center;

  .icon {
    width: 14px;
    height: 14px;
    margin-right: 4px;
  }
}
</style>
