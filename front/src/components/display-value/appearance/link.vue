<script setup lang="ts">
import { DisplayType } from '../typings';

defineProps<{
  value: string | number | string[] | number[];
  displayValue: string | number | string[] | number[];
  displayOn: DisplayType['on'];
}>();
</script>

<template>
  <bk-link theme="primary">{{ displayValue }}</bk-link>
</template>

<style lang="scss" scoped></style>
