<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { ref, watchEffect } from 'vue';

const props = defineProps<{ detail?: VendorResource<NetworkInterface, 'huawei'> }>();

const columns = ref([
  {
    label: '类型',
    field: 'type',
  },
  {
    label: '内网IP',
    field: 'internalIp',
  },
  {
    label: '弹性公网IP',
    field: 'publicIp',
  },
]);

const ipConfigData = ref([]);

watchEffect(() => {
  ipConfigData.value = [
    {
      type: '私有IP地址',
      internalIp: props.detail?.private_ipv4?.join(',') || '--',
      publicIp: props.detail?.public_ipv4?.join(',') || '--',
    },
    {
      type: 'IPv6地址',
      internalIp: props.detail?.private_ipv6?.join(',') || '--',
      publicIp: props.detail?.public_ipv6?.join(',') || '--',
    },
  ];

  props.detail?.extension?.virtual_ip_list?.forEach(({ ip, elasticity_ip }) => {
    ipConfigData.value.push({
      type: '虚拟IP地址',
      internalIp: ip || '--',
      publicIp: elasticity_ip || '--',
    });
  });
});
</script>

<template>
  <bk-table row-hover="auto" :columns="columns" :data="ipConfigData" show-overflow-tooltip />
</template>
