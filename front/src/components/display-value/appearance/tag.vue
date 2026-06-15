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
  <bk-tag v-bind="attrs">{{ displayValue }}</bk-tag>
</template>

<style lang="scss" scoped></style>
