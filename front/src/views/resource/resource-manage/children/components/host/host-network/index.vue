<script lang="ts" setup>
import TcloudNetwork from './components/tcloud-network.vue';
import HuaweiNetwork from './components/huawei-network.vue';
import AzureNetwork from './components/azure-network.vue';
import GcpNetwork from './components/gcp-network.vue';
import AwsNetwork from './components/aws-network.vue';

import { PropType, ref } from 'vue';

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
  <div class="host-network-container">
    <component :is="renderComponent" :data="props.data" :filter="filter"></component>

    <!-- <bk-dialog
    v-model:is-show="showSecurityDialog"
    :title="t('绑定安全组')"
    width="1200"
    :theme="'primary'"
    :is-loading="securityBindLoading"
    @confirm="handleSecurityConfirm">
    <bk-loading
      :loading="securityLoading"
    >
      <bk-table
        class="mt20"
        row-hover="auto"
        remote-pagination
        :columns="securityColumns"
        :data="securityDatas"
        :pagination="securityPagination"
        @selection-change="handleSelectionChange"
        @page-limit-change="securityHandlePageChange"
        @page-value-change="securityHandlePageSizeChange"
      />
    </bk-loading>
  </bk-dialog> -->
  </div>
</template>

<style lang="scss" scoped>
.host-network-container {
  :deep(.bk-table) {
    max-height: 100% !important;
  }
}
</style>
