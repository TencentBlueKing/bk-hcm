<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { h, ref, watchEffect } from 'vue';

const props = defineProps<{ detail?: VendorResource<NetworkInterface, 'azure'> }>();

const columns = ref([
  {
    label: () => h('b', '应用的 DNS 服务器'),
    field: 'name',
  },
]);

const data = ref([]);

watchEffect(() => {
  data.value = props.detail.extension?.dns_settings?.applied_dns_servers?.map((item) => ({ name: item }));
});
</script>

<template>
  <bk-table row-hover="auto" :columns="columns" :data="data" show-overflow-tooltip />
</template>
