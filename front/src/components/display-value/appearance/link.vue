<script setup lang="ts">
import { useAttrs } from 'vue';
import { DisplayType } from '../typings';

defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  displayOn: DisplayType['on'];
}>();

const attrs = useAttrs();
</script>

<template>
  <bk-link class="link" theme="primary" v-bind="attrs">{{ displayValue }}</bk-link>
</template>

<style lang="scss" scoped>
.link {
  font-size: 12px;
}
</style>
