<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import VPCCidr from '../components/vpc/vpc-cidr.vue';
import VPCRoute from '../components/vpc/vpc-route.vue';
import VPCSubnet from '../components/vpc/vpc-subnet.vue';
import bus from '@/common/bus';

import {
  ref,
  inject,
  computed,
} from 'vue';
import {
  InfoBox,
  Message,
} from 'bkui-vue';
import {
  useRoute,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';
import useDetail from '../../hooks/use-detail';
import {
  useResourceStore,
} from '@/store/resource';
import { useRegionsStore } from '@/store/useRegionsStore';
import { VendorEnum } from '@/common/constant';

const { getRegionName } = useRegionsStore();

const VPCFields: any = ref([
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '云资源 ID',
    prop: 'cloud_id',
    render(cell = '') {
      const index = cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
      const value = cell.slice(index);
      return value;
    },
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
    name: '云区域',
    prop: 'bk_cloud_id',
    render(cell: number) {
      return cell <= 0 ? '暂未绑定' : cell;
    },
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
const VPCTabs = ref([
  {
    name: 'CIDR',
    value: 'cidr',
  },
  {
    name: '子网',
    value: 'subnet',
  },
  {
    name: '路由',
    value: 'route',
  },
]);

const authVerifyData: any = inject('authVerifyData');
const isResourcePage: any = inject('isResourcePage');

const actionName = computed(() => {   // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

const resourceStore = useResourceStore();
const route = useRoute();
const {
  t,
} = useI18n();

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const {
  loading,
  detail,
} = useDetail(
  'vpcs',
  route.query.id as string,
  (detail: any) => {
    switch (detail.vendor) {
      case 'tcloud':
        VPCFields.value.push(...[
          {
            name: '默认私有网络',
            prop: 'is_default',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: 'DNS服务器',
            prop: 'dns_server_set',
            tipsContent: '腾讯云默认DNS为：183.60.83.19，183.60.82.98，如果不使用腾讯云默认DNS，将无法使用内部服务，如windows激活、NTP、YUM等',
          },
          {
            name: '本地域名',
            prop: 'domain_name',
            tipsContent: '本地域名(Domain Name), 即VPC内主机的域名后缀',
          },
          {
            name: '组播',
            prop: 'enable_multicast',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.TCLOUD, val)
          },
        ]);
        break;
      case 'aws':
        VPCFields.value.push(...[
          {
            name: '状态',
            prop: 'state',
          },
          {
            name: 'DNS 主机名',
            prop: 'enable_dns_hostnames',
            render(val: boolean) {
              return val ? '已启用' : '未启用';
            },
          },
          {
            name: 'DNS 解析',
            prop: 'enable_dns_support',
            render(val: boolean) {
              return val ? '已启用' : '未启用';
            },
          },
          {
            name: '租期',
            prop: 'instance_tenancy',
          },
          {
            name: '默认VPC',
            prop: 'is_default',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.AWS, val)
          },
        ]);
        break;
      case 'azure':
        VPCFields.value.push(...[
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.AZURE, val)
          },
          {
            name: 'DNS服务器',
            prop: 'dns_servers',
            render(val: any) {
              return val ? val : 'Azure提供的DNS服务';
            },
          },
        ]);
        VPCTabs.value.pop();
        break;
      case 'gcp':
        VPCFields.value.push(...[
          {
            name: '是否默认创建子网',
            prop: 'auto_create_subnetworks',
            render(val: boolean) {
              return val ? '是' : '否';
            },
          },
          {
            name: '动态路由模式',
            prop: 'routing_mode',
          },
          {
            name: 'VPC 网络 ULA 内部 IPv6 范围',
            prop: 'enable_ula_internal_ipv6',
            render(val: boolean) {
              return val ? '已启用' : '未启用';
            },
          },
          {
            name: '最大传输单元',
            prop: 'mtu',
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.GCP, val)
          },
        ]);
        VPCTabs.value.shift();
        break;
      case 'huawei':
        VPCFields.value.push(...[
          {
            name: '状态',
            prop: 'status',
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.HUAWEI, val)
          },
        ]);
        break;
    }
  },
);

const isBindBusiness = computed(() => {
  return detail.value.bk_biz_id !== -1 && isResourcePage.value;
});

const handleDeleteVpc = (data: any) => {
  const vpcIds = [data.id];
  const getRelateNum = (type: string, field = 'vpc_id', op = 'in') => {
    return resourceStore
      .list(
        {
          page: {
            count: true,
          },
          filter: {
            op: 'and',
            rules: [{
              field,
              op,
              value: vpcIds,
            }],
          },
        },
        type,
      );
  };
  Promise
    .all([
      getRelateNum('cvms', 'vpc_ids', 'json_overlaps'),
      getRelateNum('subnets'),
      getRelateNum('route_tables'),
      getRelateNum('network_interfaces'),
    ])
    .then(([cvmsResult, subnetsResult, routeResult, networkResult]: any) => {
      if (cvmsResult?.data?.count || subnetsResult?.data?.count || routeResult?.data?.count || networkResult?.data?.count) {
        const getMessage = (result: any, name: string) => {
          if (result?.data?.count) {
            return `${result?.data?.count}个${name}，`;
          }
          return '';
        };
        Message({
          theme: 'error',
          message: `该VPC（name：${data.name}，id：${data.id}）关联${getMessage(cvmsResult, 'CVM')}${getMessage(subnetsResult, '子网')}${getMessage(routeResult, '路由表')}${getMessage(networkResult, '网络接口')}不能删除`,
        });
      } else {
        InfoBox({
          title: '请确认是否删除',
          subTitle: `将删除【${data.name}】`,
          theme: 'danger',
          headerAlign: 'center',
          footerAlign: 'center',
          contentAlign: 'center',
          onConfirm() {
            resourceStore
              .delete(
                'vpcs',
                data.id,
              )
              .then(() => {
                Message({
                  theme: 'success',
                  message: '删除成功',
                });
              });
          },
        });
      }
    });
};
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      VPC：（{{ detail.id }}）
      <template #right>
        <div @click="showAuthDialog(actionName)">
          <bk-button
            class="w100 ml10"
            theme="primary"
            :disabled="isBindBusiness || !authVerifyData?.permissionAction[actionName]"
            @click="handleDeleteVpc(detail)"
          >
            {{ t('删除') }}
          </bk-button>
        </div>
      </template>
    </detail-header>

    <detail-info
      :detail="detail"
      :fields="VPCFields"
    />

    <detail-tab
      :tabs="VPCTabs"
    >
      <template #default="type">
        <VPCCidr v-if="type === 'cidr'" :detail="detail" />
        <VPCSubnet v-if="type === 'subnet'" :detail="detail" />
        <VPCRoute v-if="type === 'route'" :detail="detail" />
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
