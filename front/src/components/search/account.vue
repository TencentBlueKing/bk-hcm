<script setup lang="ts">
import { useAttrs } from 'vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
defineOptions({ name: 'hcm-search-account' });

const model = defineModel<string>();

const attrs = useAttrs();
</script>

<template>
  <account-selector v-model="model" :auto-select-single="false" v-bind="attrs"></account-selector>
</template>
