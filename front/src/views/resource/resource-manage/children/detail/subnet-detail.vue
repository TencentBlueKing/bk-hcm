<script lang="ts" setup>
import type { Subnet, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import { CloudType } from '@/typings/account';

import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
import SubnetRoute from '../../children/components/subnet/subnet-route.vue';
import { useResourceDetailAuth, useResourceDetailRoute } from '../../hooks/use-resource-detail-route';

import { ref, inject, computed, watchEffect } from 'vue';
import { InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useResourceDetail } from '../../hooks/use-detail';
import { useResourceStore } from '@/store/resource';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import router from '@/router';
import routerAction from '@/router/utils/action';
import { MENU_RESOURCE_ACCOUNT_DETAIL } from '@/constants/menu-symbol';
import { timeFormatter } from '@/common/util';
import { VendorEnum } from '@/common/constant';
import { FieldList } from '../../common/info-list/types';

const tcloudField = extensionPath<VendorResource<Subnet, 'tcloud'>>();
const awsField = extensionPath<VendorResource<Subnet, 'aws'>>();
const azureField = extensionPath<VendorResource<Subnet, 'azure'>>();
const gcpField = extensionPath<VendorResource<Subnet, 'gcp'>>();
const huaweiField = extensionPath<VendorResource<Subnet, 'huawei'>>();

const { getNameFromBusinessMap } = useBusinessMapStore();
const { setTitle } = useBreadcrumb();

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

const settingFields = ref<FieldList>([
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
    txtBtn(accountId: string) {
      routerAction.redirect({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId } }, { history: true });
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
    render(cell: keyof typeof CloudType) {
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

const isResourcePage: any = inject('isResourcePage');

const resourceStore = useResourceStore();
const { detailId } = useResourceDetailRoute();
const { authPending, updateSign } = useResourceDetailAuth(computed(() => detail.value.account_id));

const isBindBusiness = computed(() => {
  return detail.value.bk_biz_id !== -1 && isResourcePage.value;
});

const { getRegionName } = useRegionsStore();

const { loading, detail } = useResourceDetail<'subnets'>('subnets', detailId.value, (detail) => {
  if (detail.vendor !== 'gcp') {
    resourceStore.countSubnetIps(detailId.value).then((res: any) => {
      detail.ipv4_nums = res?.data?.available_ip_count || 0;
    });
  }
  switch (detail.vendor) {
    case 'tcloud':
      settingFields.value.push(
        ...[
          {
            name: '是否默认子网',
            prop: tcloudField('extension.is_default'),
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
            prop: tcloudField('extension.cloud_network_acl_id'),
          },
        ],
      );
      break;
    case 'aws':
      settingFields.value.push(
        ...[
          {
            name: '状态',
            prop: awsField('extension.state'),
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
            prop: awsField('extension.map_public_ip_on_launch'),
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '是否默认子网',
            prop: awsField('extension.is_default'),
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '自动分配IPv6地址',
            prop: awsField('extension.assign_ipv6_address_on_creation'),
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '主机名称类型',
            prop: awsField('extension.hostname_type'),
          },
        ],
      );
      break;
    case 'azure':
      settingFields.value.push(
        ...[
          {
            name: 'NAT网关',
            prop: azureField('extension.nat_gateway'),
          },
          {
            name: '网络安全组',
            prop: azureField('extension.cloud_security_group_id'),
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
            prop: gcpField('extension.stack_type'),
          },
          {
            name: 'IPv6权限类型',
            prop: gcpField('extension.ipv6_access_type'),
          },
          {
            name: '网关',
            prop: gcpField('extension.gateway_address'),
          },
          {
            name: '专用Google访问通道',
            prop: gcpField('extension.private_ip_google_access'),
            render(val: boolean) {
              return val ? '启用' : '关闭';
            },
          },
          {
            name: '流日志',
            prop: gcpField('extension.enable_flow_logs'),
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
            prop: huaweiField('extension.status'),
          },
          {
            name: '地域',
            prop: 'region',
            render: () => getRegionName(detail.vendor, detail.region),
          },
          {
            name: 'DHCP',
            prop: huaweiField('extension.dhcp_enable'),
            render(val: boolean) {
              return val ? '启用' : '关闭';
            },
          },
          {
            name: '网关',
            prop: huaweiField('extension.gateway_ip'),
          },
          {
            name: 'DNS服务器地址',
            prop: huaweiField('extension.dns_list'),
          },
          {
            name: 'NTP服务器地址',
            prop: huaweiField('extension.ntp_addresses'),
          },
        ],
      );
      break;
  }
});

watchEffect(() => {
  if (detail.value?.id) {
    setTitle(`子网：ID（${detail.value.id}）`);
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
</script>

<template>
  <Teleport defer to="#breadcrumbExtra">
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        v-if="isResourcePage"
        v-bk-tooltips="{
          content: '该子网已分配到业务，仅可在业务下操作',
          disabled: !isBindBusiness || noPerm,
        }"
        theme="primary"
        :disabled="isBindBusiness || noPerm || authPending"
        @click="handleDeleteSubnet(detail)"
      >
        {{ t('删除') }}
      </bk-button>
      <!-- 业务下沿用原判定：有权限时置灰（PR 同） -->
      <bk-button
        v-else
        v-bk-tooltips="{
          content: '该子网正在使用中，不能删除',
          disabled: noPerm,
        }"
        theme="primary"
        :disabled="!noPerm"
        @click="handleDeleteSubnet(detail)"
      >
        {{ t('删除') }}
      </bk-button>
    </hcm-auth>
  </Teleport>

  <div class="detail-content-wrap">
    <detail-tab :tabs="hostTabs">
      <template #default="type">
        <bk-loading :loading="loading">
          <detail-info
            v-if="type === 'detail'"
            :fields="settingFields"
            :detail="detail"
            :label-width="[VendorEnum.AWS, VendorEnum.GCP].includes(detail.vendor) ? '180px' : '120px'"
            global-copyable
          />
          <subnet-route v-if="type === 'network'" :detail="detail" />
        </bk-loading>
      </template>
    </detail-tab>
  </div>
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
