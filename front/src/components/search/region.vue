<script setup lang="ts">
import { useAttrs } from 'vue';
import { VendorEnum } from '@/common/constant';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector.vue';
defineOptions({ name: 'hcm-search-region' });

const model = defineModel<string>();
defineProps<{ vendor: VendorEnum }>();
const attrs = useAttrs();
</script>

<template>
  <region-selector v-model="model" :vendor="vendor" v-bind="attrs"></region-selector>
</template>
