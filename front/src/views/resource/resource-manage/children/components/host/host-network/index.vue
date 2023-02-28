<script lang="ts" setup>
import TcloudNetwork from './components/tcloud-network.vue';
import HuaweiNetwork from './components/huawei-network.vue';
import AzureNetwork from './components/azure-network.vue';
import GcpNetwork from './components/gcp-network.vue';
import AwsNetwork from './components/aws-network.vue';

// import {
//   useRoute,
// } from 'vue-router';

// const route = useRoute();

const componentMap = {
  tcloud: TcloudNetwork,
  huawei: HuaweiNetwork,
  azure: AzureNetwork,
  gcp: GcpNetwork,
  aws: AwsNetwork,
};

// const renderComponent = componentMap[route.params.type as string];
const renderComponent = componentMap.aws;
</script>

<template>
  <component :is="renderComponent"></component>
</template>
