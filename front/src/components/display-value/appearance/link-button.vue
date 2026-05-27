<script setup lang="ts">
import { useAttrs } from 'vue';
import { Share } from 'bkui-vue/lib/icon';
import { DisplayType } from '../typings';

defineOptions({ inheritAttrs: false });

defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  displayOn: DisplayType['on'];
  isIcon?: boolean;
}>();

const attrs = useAttrs();
</script>

<template>
  <bk-button text theme="primary" v-bind="attrs" v-if="!isIcon">
    {{ displayValue }}
  </bk-button>
  <div class="icon-link-button-container" v-else>
    <span>{{ displayValue }}</span>
    <bk-button text theme="primary" v-bind="attrs">
      <Share class="icon" />
    </bk-button>
  </div>
</template>

<style lang="scss" scoped>
.icon-link-button-container {
  display: flex;
  align-items: center;
  gap: 6px;

  .icon {
    font-size: 12px;
  }
}
</style>
