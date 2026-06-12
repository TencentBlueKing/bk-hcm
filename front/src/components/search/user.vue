<script setup lang="ts">
import { useAttrs } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';

defineOptions({ name: 'hcm-search-user' });

const model = defineModel<string | string[]>();

const attrs = useAttrs();
</script>

<template>
  <user-selector v-model="model" :allow-create="true" :multiple="true" v-bind="attrs" />
</template>
