<script setup lang="ts">
import { useAttrs } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';

defineOptions({ name: 'hcm-form-user' });

const model = defineModel<string | string[]>();

const attrs = useAttrs();
</script>

<template>
  <user-selector v-model="model" :allow-create="false" :multiple="true" v-bind="attrs" />
</template>
