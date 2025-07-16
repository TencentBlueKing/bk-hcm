<script setup lang="ts">
import { useSlots } from 'vue';
import CopyToClipboard, { type ICopyToClipboardProps } from './index.vue';

const props = defineProps<ICopyToClipboardProps>();

const slots = useSlots();
</script>

<template>
  <div class="hover-copy">
    <slot v-if="slots?.default"></slot>
    <span v-else>{{ content || '--' }}</span>
    <copy-to-clipboard v-if="content" class="copy-icon" v-bind="props" />
  </div>
</template>

<style scoped lang="scss">
.hover-copy {
  display: inline-flex;
  align-items: center;
  gap: 4px;

  .copy-icon {
    opacity: 0;
  }

  &:hover {
    .copy-icon {
      opacity: 1;
    }
  }
}
</style>
