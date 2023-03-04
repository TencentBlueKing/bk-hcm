<script lang="ts" setup>
import TcloudNetwork from './components/tcloud-network.vue';
import HuaweiNetwork from './components/huawei-network.vue';
import AzureNetwork from './components/azure-network.vue';
import GcpNetwork from './components/gcp-network.vue';
import AwsNetwork from './components/aws-network.vue';

import {
  PropType,
  ref,
} from 'vue';

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
  type: {
    type: String,
  },
});

const componentMap = {
  tcloud: TcloudNetwork,
  huawei: HuaweiNetwork,
  azure: AzureNetwork,
  gcp: GcpNetwork,
  aws: AwsNetwork,
};

// const renderComponent = componentMap[route.params.type as string];
const renderComponent = componentMap[props.type];
const filter = ref({ op: 'and', rules: [] });


</script>

<template>
  <component :is="renderComponent" :data="props.data" :filter="filter"></component>
</template>
