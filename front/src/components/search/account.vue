<script setup lang="ts">
import AccountSelector from '@/components/account-selector/index-new.vue';
defineOptions({ name: 'hcm-search-account' });

const model = defineModel<string>();
</script>

<template>
  <account-selector v-model="model"></account-selector>
</template>
