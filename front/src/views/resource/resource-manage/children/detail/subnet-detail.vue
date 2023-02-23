<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
import SubnetRoute from '../../children/components/subnet/subnet-route.vue';

import {
  ref,
  onBeforeMount
} from 'vue';
import {
  useRoute,
} from 'vue-router';
import {
  InfoBox,
} from 'bkui-vue';
import {
  useI18n,
} from 'vue-i18n';
import useDetail from '../../hooks/use-detail';
import {
  useResourceStore,
} from '@/store/resource';

const hostTabs = ref<any[]>([
  {
    name: '基本信息',
    value: 'detail',
  },
  {
    name: '路由策略',
    value: 'network',
  },
]);

const settingFields = ref<any[]>([
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '云资源 ID',
    prop: 'cloud_id',
  },
  {
    name: '资源名称',
    prop: 'name',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: string) {
      return CloudType[cell] || '--';
    },
  },
  {
    name: '所属 VPC',
    prop: 'vpc_id',
  },
  {
    name: 'IPv4 CIDR',
    prop: 'ipv4_cidr',
  },
  {
    name: '可用 IPv4 地址数',
    prop: 'ipv4_nums',
  },
  {
    name: 'IPv6 CIDR',
    prop: 'ipv6_cidr',
  },
  {
    name: '创建时间',
    prop: 'created_at',
  },
  {
    name: '备注',
    type: 'textarea',
    prop: 'memo',
    // edit: true,
  },
]);

const {
  t,
} = useI18n();

const resourceStore = useResourceStore();
const route = useRoute();

const {
  loading,
  detail,
} = useDetail(
  'subnets',
  route.query.id as string,
  (detail: any) => {
    switch (detail.vendor) {
      case 'tcloud':
        settingFields.value.push(...[
          {
            name: '是否默认子网',
            prop: 'is_default',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '地域',
            prop: 'region',
          },
          {
            name: '可用区',
            prop: 'zone',
          },
          {
            name: '关联ACL',
            prop: 'network_acl_id',
          },
        ]);
        break;
      case 'aws':
        settingFields.value.push(...[
          {
            name: '状态',
            prop: 'state',
          },
          {
            name: '地域',
            prop: 'region',
          },
          {
            name: '可用区',
            prop: 'zone',
          },
          {
            name: '自动分配公有 IPv4 地址',
            prop: 'map_public_ip_on_launch',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '是否默认子网',
            prop: 'is_default',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '自动分配 IPv6 地址',
            prop: 'assign_ipv6_address_on_creation',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '主机名称类型',
            prop: 'hostname_type',
          },
        ]);
        break;
      case 'azure':
        settingFields.value.push(...[
          {
            name: 'NAT网关',
            prop: 'nat_gateway',
          },
          {
            name: '网络安全组',
            prop: 'network_security_group',
          },
        ]);
        break;
      case 'gcp':
        settingFields.value.push(...[
          {
            name: '地域',
            prop: 'region',
          },
          {
            name: 'IP 栈类型',
            prop: 'stack_type',
          },
          {
            name: 'IPv6 权限类型',
            prop: 'ipv6_access_type',
          },
          {
            name: '网关',
            prop: 'gateway_address',
          },
          {
            name: '专用 Google 访问通道',
            prop: 'private_ip_google_access',
            render(val: boolean) {
              return val ? '启用' : '关闭';
            },
          },
          {
            name: '流日志',
            prop: 'enable_flow_logs',
            render(val: boolean) {
              return val ? '启用' : '关闭';
            },
          },
        ]);
        hostTabs.value.pop();
        break;
      case 'huawei':
        settingFields.value.push(...[
          {
            name: '状态',
            prop: 'status',
          },
          {
            name: '地域',
            prop: 'region',
          },
          {
            name: 'DHCP',
            prop: 'dhcp_enable',
            render(val: boolean) {
              return val ? '启用' : '关闭';
            },
          },
          {
            name: '网关',
            prop: 'gateway_ip',
          },
          {
            name: 'DNS 服务器地址',
            prop: 'dns_list',
          },
          {
            name: 'NTP 服务器地址',
            prop: 'ntp_addresses',
          },
        ]);
        break;
    }
  },
);

const handleShowDelete = () => {
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${detail.value.name}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      return resourceStore
        .deleteBatch(
          'subnets',
          {
            ids: [detail.value.id],
          },
        );
    },
  });
};

onBeforeMount(() => {
  resourceStore
    .countSubnetIps(route.query.id as string)
    .then((res: any) => {
      detail.value.ipv4_nums = res?.data?.available_ipv4_count || 0
    })
})
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      子网：ID（xxx）
      <template #right>
        <bk-button
          class="w100 ml10"
          theme="primary"
          @click="handleShowDelete"
        >
          {{ t('删除') }}
        </bk-button>
      </template>
    </detail-header>

    <detail-tab
      :tabs="hostTabs"
    >
      <template #default="type">
        <detail-info
          v-if="type === 'detail'"
          :fields="settingFields"
          :detail="detail"
        />
        <subnet-route
          v-if="type === 'network'"
        />
      </template>
    </detail-tab>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
