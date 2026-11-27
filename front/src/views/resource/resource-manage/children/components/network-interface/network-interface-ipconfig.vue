<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { ref, watch } from 'vue';
import { FieldList } from '../../../common/info-list/types';

const props = defineProps<{ detail?: VendorResource<NetworkInterface, 'azure'> }>();

const fieldPath = extensionPath<VendorResource<NetworkInterface, 'azure'>>();

const fields = ref<FieldList>([
  {
    name: 'IP转发',
    prop: fieldPath('extension.enable_ip_forwarding'),
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
    prop: fieldPath('extension.cloud_gateway_load_balancer_id'),
  },
  {
    name: '虚拟网络(VPC)',
    prop: 'cloud_vpc_id',
  },
]);

const columns = ref([]);
const ipConfigData = ref([]);

watch(
  () => props.detail,
  (detail) => {
    if (detail.vendor === 'azure') {
      ipConfigData.value = (detail.extension?.ip_configurations ?? []).map((data) => ({
        ...data,
        ...data.properties,
        publicIPAddress: data.properties?.public_ip_address?.properties?.ip_address || '--',
      }));

      columns.value = [
        {
          label: '名称',
          field: 'name',
        },
        {
          label: 'IP版本',
          field: 'private_ip_address_version',
        },
        {
          label: '类型',
          field: 'primary',
          render({ cell }: { cell: number }) {
            return cell ? '主要' : '辅助';
          },
        },
        {
          label: '专用IP地址',
          field: 'private_ip_address',
        },
        {
          label: '公共IP地址',
          field: 'publicIPAddress',
        },
      ];
    }
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>

<template>
  <div class="ipconfig">
    <detail-info :detail="detail" :fields="fields" global-copyable />
    <bk-table class="mt20" row-hover="auto" :columns="columns" :data="ipConfigData" show-overflow-tooltip />
  </div>
</template>

<style lang="scss" scoped>
.ipconfig {
  :deep(.detail-info-main) {
    height: auto;
  }
}
</style>
