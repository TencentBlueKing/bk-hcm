<script lang="ts" setup>
import TcloudInfo from './components/tcloud-info.vue';
import AwsInfo from './components/aws-info.vue';
import GcpInfo from './components/gcp-info.vue';
import AzureInfo from './components/azure-info.vue';
import HuaweiInfo from './components/huawei-info.vue';
// import HuaweiNetwork from './components/huawei-network.vue';
// import AzureNetwork from './components/azure-network.vue';
// import GcpNetwork from './components/gcp-network.vue';

import {
  PropType,
} from 'vue';

// import {
//   useRoute,
// } from 'vue-router';

// const route = useRoute();

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
  type: {
    type: String,
  },
});

const componentMap = {
  tcloud: TcloudInfo,
  huawei: HuaweiInfo,
  azure: AzureInfo,
  gcp: GcpInfo,
  aws: AwsInfo,
};

// const renderComponent = componentMap[route.params.type as string];
console.log('props.data.vendor', props.type);
const renderComponent = componentMap[props.type];
console.log('renderComponent', renderComponent);
</script>

<template>
  <component :is="renderComponent" :data="props.data"></component>
</template>
