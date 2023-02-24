<script lang="ts" setup>
import { ref, watchEffect } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

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
      internalIp: props.detail?.internal_ip,
      publicIp: props.detail?.addresses ? ([
        props.detail?.addresses?.public_ip_address || props.detail?.addresses?.public_ipv6_address,
        props.detail?.addresses?.bandwidth_type,
        `${props.detail?.addresses?.bandwidth_size} M/s`,
      ]).join(' | ') : '--',
    },
    {
      type: 'IPv6地址',
      internalIp: props.detail?.ipv6 || '--',
      publicIp: '--',
    },
  ];

  props.detail?.virtual_ip_list?.forEach(({ ip, elasticity_ip }: { ip: string, elasticity_ip: string }) => {
    ipConfigData.value.push({
      type: '虚拟IP地址',
      internalIp: ip || '--',
      publicIp: elasticity_ip || '--',
    });
  });
});
</script>

<template>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="ipConfigData"
  />
</template>
