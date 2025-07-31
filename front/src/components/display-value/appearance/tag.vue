<script setup lang="ts">
import { DisplayType } from '../typings';

defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  displayOn: DisplayType['on'];
}>();
</script>

<template>
  <bk-tag theme="primary">{{ displayValue }}</bk-tag>
</template>

<style lang="scss" scoped></style>
