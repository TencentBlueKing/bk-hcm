<script setup lang="ts">
import { computed } from 'vue';
import { formatJSON } from '@/utils/common';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { DisplayType } from './typings';

const props = withDefaults(defineProps<{ value: any; copyable?: boolean; display?: DisplayType }>(), {
  copyable: true,
});

const displayOn = computed(() => props.display?.on);
const displayValue = computed(() => formatJSON(props.value, displayOn.value === 'info' ? 2 : 0));
</script>

<template>
  <template v-if="displayOn === 'info'">
    <div class="json-value-info">
      <pre>{{ displayValue }}</pre>
      <copy-to-clipboard :content="displayValue" v-if="copyable" class="copy-btn" />
    </div>
  </template>
  <template v-else>
    <bk-overflow-title class="full-width" resizeable type="tips" v-if="display?.showOverflowTooltip">
      {{ displayValue }}
    </bk-overflow-title>
    <span v-else>{{ displayValue }}</span>
  </template>
</template>

<style lang="scss" scoped>
.json-value-info {
  width: 100%;
  background: #f5f7fa;
  border-radius: 2px;
  padding: 16px;
  font-size: 12px;
  line-height: 20px;
  overflow-x: auto;
  color: #000;

  .copy-btn {
    position: absolute;
    right: 16px;
    top: 16px;
  }
}
</style>
