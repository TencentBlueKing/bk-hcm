<script setup lang="ts">
import { computed } from 'vue';
import { DisplayType } from './typings';

const props = defineProps<{ value: string | number; display: DisplayType }>();

const displayValue = computed(() => props.value || '--');
</script>

<template>
  <bk-overflow-title resizeable type="tips" v-if="display?.showOverflowTooltip">{{ displayValue }}</bk-overflow-title>
  <span v-else>{{ displayValue }}</span>
</template>
