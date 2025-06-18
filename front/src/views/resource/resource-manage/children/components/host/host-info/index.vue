<script lang="ts" setup>
import TcloudInfo from './components/tcloud-info.vue';
import AwsInfo from './components/aws-info.vue';
import GcpInfo from './components/gcp-info.vue';
import AzureInfo from './components/azure-info.vue';
import HuaweiInfo from './components/huawei-info.vue';
import OtherInfo from './components/other-info.vue';
import { VendorEnum } from '@/common/constant';

const props = defineProps<{ data: any; type: VendorEnum }>();

const componentMap: { [key in VendorEnum]?: any } = {
  tcloud: TcloudInfo,
  huawei: HuaweiInfo,
  azure: AzureInfo,
  gcp: GcpInfo,
  aws: AwsInfo,
  other: OtherInfo,
};

const renderComponent = componentMap[props.type];
</script>

<template>
  <component :is="renderComponent" :data="props.data"></component>
</template>
<style lang="scss" scoped>
.f-right {
  float: right;
}
</style>
