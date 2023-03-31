<script lang="ts" setup>
import {
  defineProps,
  ref,
  watch,
} from 'vue';
import DetailList from '../../../common/info/detail-info';
import DetailTab from '../../../common/tab/detail-tab';

const props = defineProps({
  detail: Object
});

const baseTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
const bindTabs = [
  {
    name: '绑定信息',
    value: 'detail',
  },
];
const otherTabs = [
  {
    name: '其他信息',
    value: 'detail',
  },
];
const huaweiTabs = [
  {
    name: '带宽',
    value: 'detail',
  },
];

const baseInfo = [
  {
    name: 'EIP名称',
    prop: 'name',
  },
  {
    name: 'ID',
    prop: 'id',
  },
  {
    name: '云资源ID',
    prop: 'cloud_id',
  },
  {
    name: 'IP地址',
    prop: 'public_ip',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '地域',
    prop: 'region',
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
  },
  {
    name: '创建时间',
    prop: 'created_at',
  },
];

const bindInfo = ref([
  {
    name: '绑定的资源类型',
    prop: 'instance_type',
  },
  {
    name: '绑定资源实例',
    prop: 'instance_id',
  },
  {
    name: '状态',
    prop: 'status',
  },
]);

const otherInfo = ref([]);

const bandInfo = [
  {
    name: '带宽名称',
    prop: 'bandwidth_name',
  },
  {
    name: '带宽ID',
    prop: 'bandwidth_id',
  },
  // {
  //   name: '计费模式',
  //   prop: '',
  // },
  // {
  //   name: '计费方式',
  //   prop: '',
  // },
  {
    name: '带宽大小 (Mbit/s)',
    prop: 'bandwidth_size',
  },
  {
    name: '带宽类型',
    prop: 'bandwidth_share_type',
  }
]

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        otherInfo.value = [
          {
            name: '线路类型',
            prop: 'internet_service_provider',
          },
          {
            name: '带宽上限',
            prop: 'bandwidth',
          },
          {
            name: '加速地区',
            prop: '',
          },
          {
            name: '计费模式',
            prop: 'internet_charge_type',
          },
        ];
        baseInfo.splice(7, 1);
        break;
      case 'aws':
        otherInfo.value = [
          {
            name: '范围',
            prop: 'domain',
          },
          {
            name: '地址池',
            prop: 'public_ipv4_pool',
          },
          {
            name: '内网IP',
            prop: 'private_ip_address',
          },
          {
            name: '网络接口',
            prop: 'network_interface_id',
          },
          {
            name: 'NAT网关ID',
            prop: '',
          },
          {
            name: '公网DNS',
            prop: '',
          },
          {
            name: '反向DNS解析',
            prop: '',
          },
        ];
        break;
      case 'gcp':
        otherInfo.value = [
          {
            name: '权限类型',
            prop: 'address_type',
          },
          {
            name: '类型',
            value: '静态'
          },
          {
            name: '网络层',
            prop: 'network_tier',
          },
        ];
        break;
      case 'azure':
        otherInfo.value = [
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
          {
            name: '位置',
            prop: 'location',
          },
          {
            name: 'SKU',
            prop: 'sku',
          },
          {
            name: '层',
            prop: 'sku_tier',
          },
          {
            name: 'DNS 名称',
            prop: '',
          },
        ];
        break;
      case 'huawei':
        otherInfo.value = [
          {
            name: '企业项目',
            prop: 'enterprise_project_id',
          },
          {
            name: '子网',
            prop: '',
          },
          {
            name: '已绑定网卡',
            prop: '',
          },
          {
            name: '类型',
            prop: 'type',
          },
        ];
        break;
    }
  }
)
</script>

<template>
  <detail-tab
    :tabs="baseTabs"
    class="auto-tab"
  >
    <template #default>
      <detail-list
        class="mt20"
        :fields="baseInfo"
        :detail="detail"
      ></detail-list>
    </template>
  </detail-tab>

  <detail-tab
    :tabs="bindTabs"
    class="auto-tab"
  >
    <template #default>
      <detail-list
        class="mt20"
        :fields="bindInfo"
        :detail="detail"
      ></detail-list>
    </template>
  </detail-tab>

  <detail-tab
    :tabs="otherTabs"
    class="auto-tab"
  >
    <template #default>
      <detail-list
        class="mt20"
        :fields="otherInfo"
        :detail="detail"
      ></detail-list>
    </template>
  </detail-tab>

  <detail-tab
    v-if="detail.vendor === 'huawei'"
    :tabs="huaweiTabs"
    class="auto-tab"
  >
    <template #default>
      <detail-list
        class="mt20"
        :fields="bandInfo"
        :detail="detail"
      ></detail-list>
    </template>
  </detail-tab>
</template>

<style lang="scss">
  .auto-tab {
    .bk-tab-content, .detail-info-main, .bk-tab-panel {
      height: auto !important;
    }
  }
</style>
