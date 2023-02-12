<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { ref, watch } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const fields = ref([
  {
    name: 'IP转发',
    prop: 'enable_ip_forwarding',
    render(val: boolean) {
      return val ? '是' : '否';
    },
  },
  {
    name: '子网',
    prop: 'cloud_subnet_id',
  },
  {
    name: '网关负载均衡器',
    prop: 'gatewayLoadBalancerId',
  },
  {
    name: '虚拟网络(VPC)',
    prop: 'virtual_network',
  },
]);


const columns = ref([]);
const ipConfigData = ref([]);

watch(
  () => props.detail,
  (detail) => {
    switch (detail.vendor) {
      case 'tcloud':
      case 'azure':
      case 'huawei':
        ipConfigData.value = detail.ip_configurations.map((data: any) => ({
          ...data,
          ...data.properties,
          publicIPAddress: data.properties.publicIPAddress.name,
        }));

        columns.value = [
          {
            label: '名称',
            field: 'name',
          },
          {
            label: 'IP版本',
            field: 'privateIPAddressVersion',
          },
          {
            label: '类型',
            field: 'type',
          },
          {
            label: '专用IP地址',
            field: 'privateIPAddress',
          },
          {
            label: '公共IP地址',
            field: 'publicIPAddress',
          },
        ];
        break;
    }
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>

<template>
  <detail-info
    class="mt20"
    :detail="detail"
    :fields="fields"
  />

  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="ipConfigData"
  />
</template>
