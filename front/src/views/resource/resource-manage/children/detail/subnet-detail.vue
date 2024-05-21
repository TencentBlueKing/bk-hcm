<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
import SubnetRoute from '../../children/components/subnet/subnet-route.vue';
import bus from '@/common/bus';

import { ref, inject, computed, onBeforeMount } from 'vue';
import { useRoute } from 'vue-router';
import { InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import useDetail from '../../hooks/use-detail';
import { useResourceStore } from '@/store/resource';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import router from '@/router';
import { timeFormatter } from '@/common/util';

const { getNameFromBusinessMap } = useBusinessMapStore();
const { whereAmI } = useWhereAmI();

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
    name: '资源ID',
    prop: 'id',
  },
  {
    name: '云资源ID',
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
      return `/#/resource/account/detail/?accountId=${val}&id=${val}`;
    },
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: string) {
      return CloudType[cell] || '--';
    },
  },
  {
    name: '所属VPC',
    prop: 'vpc_id',
  },
  {
    name: 'IPv4 CIDR',
    prop: 'ipv4_cidr',
  },
  {
    name: '可用IPv4地址数',
    prop: 'ipv4_nums',
  },
  {
    name: 'IPv6 CIDR',
    prop: 'ipv6_cidr',
  },
  {
    name: '创建时间',
    prop: 'created_at',
    render: (val: string) => timeFormatter(val),
  },
  {
    name: '备注',
    type: 'textarea',
    prop: 'memo',
    // edit: true,
  },
]);

const { t } = useI18n();

const authVerifyData: any = inject('authVerifyData');
const isResourcePage: any = inject('isResourcePage');

const actionName = computed(() => {
  // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

const resourceStore = useResourceStore();
const route = useRoute();

const isBindBusiness = computed(() => {
  return detail.value.bk_biz_id !== -1 && isResourcePage.value;
});

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const { getRegionName } = useRegionsStore();

const { loading, detail } = useDetail('subnets', route.query.id as string, (detail: any) => {
  switch (detail.vendor) {
    case 'tcloud':
      settingFields.value.push(
        ...[
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
            render: () => getRegionName(detail.vendor, detail.region),
          },
          {
            name: '可用区',
            prop: 'zone',
          },
          {
            name: '关联ACL',
            prop: 'network_acl_id',
          },
        ],
      );
      break;
    case 'aws':
      settingFields.value.push(
        ...[
          {
            name: '状态',
            prop: 'state',
          },
          {
            name: '地域',
            prop: 'region',
            render: () => getRegionName(detail.vendor, detail.region),
          },
          {
            name: '可用区',
            prop: 'zone',
          },
          {
            name: '自动分配公有IPv4地址',
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
            name: '自动分配IPv6地址',
            prop: 'assign_ipv6_address_on_creation',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '主机名称类型',
            prop: 'hostname_type',
          },
        ],
      );
      break;
    case 'azure':
      settingFields.value.push(
        ...[
          {
            name: 'NAT网关',
            prop: 'nat_gateway',
          },
          {
            name: '网络安全组',
            prop: 'network_security_group',
          },
        ],
      );
      break;
    case 'gcp':
      settingFields.value.push(
        ...[
          {
            name: '地域',
            prop: 'region',
            render: () => getRegionName(detail.vendor, detail.region),
          },
          {
            name: 'IP栈类型',
            prop: 'stack_type',
          },
          {
            name: 'IPv6权限类型',
            prop: 'ipv6_access_type',
          },
          {
            name: '网关',
            prop: 'gateway_address',
          },
          {
            name: '专用Google访问通道',
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
        ],
      );
      hostTabs.value.pop();
      break;
    case 'huawei':
      settingFields.value.push(
        ...[
          {
            name: '状态',
            prop: 'status',
          },
          {
            name: '地域',
            prop: 'region',
            render: () => getRegionName(detail.vendor, detail.region),
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
            name: 'DNS服务器地址',
            prop: 'dns_list',
          },
          {
            name: 'NTP服务器地址',
            prop: 'ntp_addresses',
          },
        ],
      );
      break;
  }
});

const handleDeleteSubnet = (data: any) => {
  const subnetIds = [data.id];
  const getRelateNum = (type: string, field = 'subnet_id', op = 'in') => {
    return resourceStore.list(
      {
        page: {
          count: true,
        },
        filter: {
          op: 'and',
          rules: [
            {
              field,
              op,
              value: subnetIds,
            },
          ],
        },
      },
      type,
    );
  };
  Promise.all([getRelateNum('cvms', 'subnet_ids', 'json_overlaps'), getRelateNum('network_interfaces')]).then(
    ([cvmsResult, networkResult]: any) => {
      if (cvmsResult?.data?.count || networkResult?.data?.count) {
        const getMessage = (result: any, name: string) => {
          if (result?.data?.count) {
            return `${result?.data?.count}个${name}，`;
          }
          return '';
        };
        Message({
          theme: 'error',
          message: `该子网（name：${data.name}，id：${data.id}）关联${getMessage(cvmsResult, 'CVM')}${getMessage(
            networkResult,
            '网络接口',
          )}不能删除`,
        });
      } else {
        InfoBox({
          title: '请确认是否删除',
          subTitle: `将删除【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
          theme: 'danger',
          headerAlign: 'center',
          footerAlign: 'center',
          contentAlign: 'center',
          extCls: 'delete-resource-infobox',
          onConfirm() {
            resourceStore
              .deleteBatch('subnets', {
                ids: [data.id],
              })
              .then(() => {
                Message({
                  theme: 'success',
                  message: '删除成功',
                });
                router.back();
              });
          },
        });
      }
    },
  );
};

onBeforeMount(() => {
  if (route.query.type === 'gcp') return;
  resourceStore.countSubnetIps(route.query.id as string).then((res: any) => {
    detail.value.ipv4_nums = res?.data?.available_ip_count || 0;
  });
});
</script>

<template>
  <bk-loading :loading="loading">
    <detail-header>
      子网：ID（{{ detail.id }}）
      <template #right>
        <div
          v-if="isResourcePage"
          v-bk-tooltips="{
            content: '该子网已分配到业务，仅可在业务下操作',
            disabled: !isBindBusiness || !authVerifyData?.permissionAction[actionName],
          }"
          @click="showAuthDialog(actionName)"
        >
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="isBindBusiness || !authVerifyData?.permissionAction[actionName]"
            @click="handleDeleteSubnet(detail)"
          >
            {{ t('删除') }}
          </bk-button>
        </div>
        <div
          v-else
          @click="showAuthDialog(actionName)"
          v-bk-tooltips="{
            content: '该子网正在使用中，不能删除',
            disabled: !authVerifyData?.permissionAction[actionName],
          }"
        >
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="authVerifyData?.permissionAction[actionName]"
            @click="handleDeleteSubnet(detail)"
          >
            {{ t('删除') }}
          </bk-button>
        </div>
      </template>
    </detail-header>

    <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
      <detail-tab :tabs="hostTabs">
        <template #default="type">
          <detail-info v-if="type === 'detail'" :fields="settingFields" :detail="detail" />
          <subnet-route v-if="type === 'network'" :detail="detail" />
        </template>
      </detail-tab>
    </div>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
:deep(.detail-info-main) {
  max-height: 100%;
  .info-list-item .item-field {
    width: 150px !important;
  }
}
</style>
