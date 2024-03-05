<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import NetworkInterfaceInfo from '../components/network-interface/network-interface-info.vue';
import NetworkInterfaceInfoGcp from '../components/network-interface/network-interface-info-gcp.vue';
import NetworkInterfaceInfoHuawei from '../components/network-interface/network-interface-info-huawei.vue';
import NetworkInterfaceIpconfig from '../components/network-interface/network-interface-ipconfig.vue';
import NetworkInterfaceIpconfigHuawei from '../components/network-interface/network-interface-ipconfig-huawei.vue';
import NetworkInterfaceDnssvr from '../components/network-interface/network-interface-dnssvr.vue';
import NetworkInterfaceNetsecgroup from '../components/network-interface/network-interface-netsecgroup.vue';

import {
  useRoute,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useDetail from '../../hooks/use-detail';
import { computed } from '@vue/runtime-core';

import {
  inject,
} from 'vue';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const route = useRoute();
const {
  t,
} = useI18n();

const isResourcePage: any = inject('isResourcePage');
const { whereAmI } = useWhereAmI();

console.log('isResourcePage', isResourcePage.value);

const {
  loading,
  detail,
} = useDetail('network_interfaces', route.query.id as string, (data: any) => {
  data.virtualNetworkSubnetId = `${data.cloud_vpc_id || '--'}/${data.cloud_subnet_id || '--'}`;
  switch (data.vendor) {
    case 'azure':
      data.gatewayLoadBalancerId = data.gateway_load_balancer?.id;
      data.associated = [];
      if (data.network_security_group?.id) {
        data.associated.push({
          id: data.network_security_group.id,
          name: data.network_security_group.id.split('/')?.pop(),
          label: '网络安全组',
        });
      }
      if (data.virtual_machine?.id) {
        data.associated.push({
          id: data.virtual_machine.id,
          name: data.virtual_machine.id.split('/')?.pop(),
          label: '虚拟机',
        });
      }
      break;
  }
});


const tabs = computed(() => {
  const list = [
    {
      name: '基本信息',
      value: 'basic',
    },
  ];
  if (detail.value.vendor === 'azure') {
    list.push(
      {
        name: 'IP信息',
        value: 'ipconfig',
      },
      {
        name: 'DNS服务器',
        value: 'dnssvr',
      },
      {
        name: '网络安全组',
        value: 'netsecgroup',
      },
    );
  }
  if (detail.value.vendor === 'huawei') {
    list.push({
      name: 'IP信息',
      value: 'ipconfig',
    });
  }

  return list;
});

</script>

<template>
  <bk-loading :loading="loading">
    <detail-header>
      {{ t('网络接口') }}：ID（{{ detail.id }}）
    </detail-header>

    <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
      <detail-tab :tabs="tabs">
        <template #default="type">
          <template v-if="detail.vendor === 'azure'">
            <network-interface-info :detail="detail" v-if="type === 'basic'"></network-interface-info>
            <network-interface-ipconfig :detail="detail" v-if="type === 'ipconfig'"></network-interface-ipconfig>
            <network-interface-dnssvr :detail="detail" v-if="type === 'dnssvr'"></network-interface-dnssvr>
            <network-interface-netsecgroup :detail="detail" v-if="type === 'netsecgroup'">
            </network-interface-netsecgroup>
          </template>
          <template v-else-if="detail.vendor === 'gcp'">
            <network-interface-info-gcp :detail="detail" :is-resource-page="isResourcePage"
                                        v-if="type === 'basic'"></network-interface-info-gcp>
          </template>
          <template v-else-if="detail.vendor === 'huawei'">
            <network-interface-info-huawei :detail="detail" v-if="type === 'basic'">
            </network-interface-info-huawei>
            <network-interface-ipconfig-huawei :detail="detail" v-if="type === 'ipconfig'">
            </network-interface-ipconfig-huawei>
          </template>
        </template>
      </detail-tab>
    </div>
  </bk-loading>
</template>

<style lang="scss" scoped>
.field-list {
  :deep(.cell-content-list) {
    line-height: normal;
    .cell-content-item {
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}
</style>
