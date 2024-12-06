<script setup lang="ts">
import { PlayShape } from 'bkui-vue/lib/icon';

defineProps<{ node: any }>();
</script>

<template>
  <PlayShape
    v-if="node.type !== 'loading'"
    :style="{
      width: '10px',
      color:
        (node.type === 'lb' && node.listenerNum === 0) ||
        (node.type === 'listener' && node.domain_num === 0) ||
        node.type === 'domain'
          ? 'transparent'
          : '#979ba5',
      transform: `${node.__attr__.isOpen ? 'rotate(90deg)' : 'rotate(0)'}`,
    }"
  />
</template>

<style scoped></style>
