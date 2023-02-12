<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import NetworkInterfaceInfo from '../components/network-interface/network-interface-info.vue';
import NetworkInterfaceIpconfig from '../components/network-interface/network-interface-ipconfig.vue';
import NetworkInterfaceDnssvr from '../components/network-interface/network-interface-dnssvr.vue';
import NetworkInterfaceNetsecgroup from '../components/network-interface/network-interface-netsecgroup.vue';

import { ref } from 'vue';
import {
  useRoute,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useDetail from '../../hooks/use-detail';


const route = useRoute();
const {
  t,
} = useI18n();

const {
  loading,
  detail,
} = useDetail(`vendors/${route.query.vendor}/network_interface`, route.query.id as string, (data: any) => {
  data.virtualNetworkSubnetId = `${data.virtual_network}${data.cloud_subnet_id}`;
  switch (data.vendor) {
    case 'azure':
      data.gatewayLoadBalancerId = data.gateway_load_balancer.id;
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

const tabs = [
  {
    name: '基本信息',
    value: 'basic',
  },
  {
    name: 'IP配置',
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
];
</script>

<template>
  <bk-loading :loading="loading">
    <detail-header>
      {{ t('网络接口') }}：ID（{{detail.id}}）
    </detail-header>

    <detail-tab
      :tabs="tabs"
    >
      <template #default="type">
        <network-interface-info :detail="detail" v-if="type === 'basic'"></network-interface-info>
        <network-interface-ipconfig :detail="detail" v-if="type === 'ipconfig'"></network-interface-ipconfig>
        <network-interface-dnssvr :detail="detail" v-if="type === 'dnssvr'"></network-interface-dnssvr>
        <network-interface-netsecgroup :detail="detail" v-if="type === 'netsecgroup'"></network-interface-netsecgroup>
      </template>
    </detail-tab>
  </bk-loading>
</template>
