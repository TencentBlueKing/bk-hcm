<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Button } from 'bkui-vue';
import { h, ref, watch } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const fields = ref([
  {
    name: '网络接口名称',
    prop: 'name',
  },
  {
    name: '网络接口ID',
    prop: 'id',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
  {
    name: '资源组',
    prop: 'resource_group_name',
  },
  {
    name: '位置',
    prop: 'location',
  },
]);

watch(
  () => props.detail,
  (detail) => {
    switch (detail.vendor) {
      case 'azure':
        fields.value.push(...[
          {
            name: '订阅',
            prop: 'subscribe_name',
          },
          {
            name: '订阅ID',
            prop: 'subscribe_id',
          },
          {
            name: '内网IPv4地址',
            prop: 'private_ipv4',
          },
          {
            name: '公网IPv4地址',
            prop: 'public_ipv4',
          },
          {
            name: '专用IP地址(IPv6)',
            prop: 'private_ipv6',
          },
          {
            name: '公用IP地址(IPv6)',
            prop: 'public_ipv6',
          },
          {
            name: '更快的网络连接',
            prop: 'enable_accelerated_networking',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '已关联到',
            prop: 'associated',
            render(val: [{ id: string, name: string, label: string }]) {
              return h('div', { style: { 'line-height': 'normal' } }, val.map(({ name, label }) => h('p', h(
                Button,
                { text: true, theme: 'primary' },
                `${name}(${label})`,
              ))));
            },
          },
          {
            name: '虚拟网络/子网',
            prop: 'virtualNetworkSubnetId',
          },
          {
            name: 'MAC地址',
            prop: 'mac_address',
          },
          {
            name: '负载均衡器',
            prop: 'gatewayLoadBalancerId',
          },
        ]);
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
</template>
