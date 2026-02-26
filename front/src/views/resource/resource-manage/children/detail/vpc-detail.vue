<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import VPCCidr from '../components/vpc/vpc-cidr.vue';
import VPCRoute from '../components/vpc/vpc-route.vue';
import VPCSubnet from '../components/vpc/vpc-subnet.vue';

import { ref, inject, computed, watch, watchEffect } from 'vue';
import { InfoBox, Message } from 'bkui-vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useDetail from '../../hooks/use-detail';
import { useResourceStore } from '@/store/resource';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { VendorEnum } from '@/common/constant';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { timeFormatter } from '@/common/util';
import { FieldList } from '../../common/info-list/types';
import { AUTH_DELETE_IAAS_RESOURCE, AUTH_BIZ_DELETE_IAAS_RESOURCE } from '@/constants/auth-symbols';

const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const { whereAmI, getBizsId } = useWhereAmI();
const bizId = computed(() => getBizsId());
const { setTitle } = useBreadcrumb();

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
const VPCFields = ref<FieldList>([
  {
    name: '资源ID',
    prop: 'id',
  },
  {
    name: '云资源ID',
    prop: 'cloud_id',
    // eslint-disable-next-line @typescript-eslint/no-inferrable-types
    render(cell: string = '') {
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
      return `/#/service/account/details/${val}`;
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
    name: '云区域',
    prop: 'bk_cloud_id',
    render(cell: number) {
      return cell <= 0 ? '暂未绑定' : cell;
    },
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

const isResourcePage: any = inject('isResourcePage');

const resourceStore = useResourceStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const { loading, detail } = useDetail('vpcs', route.params.id as string, (detail: any) => {
  switch (detail.vendor) {
    case 'tcloud':
      VPCFields.value.push(
        ...[
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
            tipsContent:
              '腾讯云默认DNS为：183.60.83.19，183.60.82.98，如果不使用腾讯云默认DNS，将无法使用内部服务，如windows激活、NTP、YUM等',
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
            render: (val: string) => getRegionName(VendorEnum.TCLOUD, val),
          },
        ],
      );
      break;
    case 'aws':
      VPCFields.value.push(
        ...[
          {
            name: '状态',
            prop: 'state',
          },
          {
            name: 'DNS主机名',
            prop: 'enable_dns_hostnames',
            render(val: boolean) {
              return val ? '已启用' : '未启用';
            },
          },
          {
            name: 'DNS解析',
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
            render: (val: string) => getRegionName(VendorEnum.AWS, val),
          },
        ],
      );
      break;
    case 'azure':
      VPCFields.value.push(
        ...[
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.AZURE, val),
          },
          {
            name: 'DNS服务器',
            prop: 'dns_servers',
            render(val: any) {
              return val.length > 0 ? val : 'Azure提供的DNS服务';
            },
          },
        ],
      );
      VPCTabs.value.pop();
      break;
    case 'gcp':
      VPCFields.value.push(
        ...[
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
            name: 'VPC网络ULA内部IPv6范围',
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
            render: (val: string) => getRegionName(VendorEnum.GCP, val),
          },
        ],
      );
      VPCTabs.value.shift();
      break;
    case 'huawei':
      VPCFields.value.push(
        ...[
          {
            name: '状态',
            prop: 'status',
          },
          {
            name: '地域',
            prop: 'region',
            render: (val: string) => getRegionName(VendorEnum.HUAWEI, val),
          },
        ],
      );
      break;
  }
});

const deleteSign = computed(() => {
  if (bizId.value) return { type: AUTH_BIZ_DELETE_IAAS_RESOURCE, relation: [bizId.value] };
  return { type: AUTH_DELETE_IAAS_RESOURCE, relation: [detail.value.account_id] };
});

watchEffect(() => {
  if (detail.value?.id) {
    setTitle(`VPC：（${detail.value.id}）`);
  }
});

const vpcRelateSubnetCount = ref(0);
watch(
  () => detail.value.id,
  (val) => {
    resourceStore
      .list(
        {
          page: { count: true },
          filter: {
            op: 'and',
            rules: [{ field: 'vpc_id', op: 'in', value: [val] }],
          },
        },
        'subnets',
      )
      .then((res: any) => {
        vpcRelateSubnetCount.value = res.data.count;
      });
  },
);

const disabledOption = computed(() => {
  if (!isResourcePage.value) return vpcRelateSubnetCount.value > 0;
  return detail.value?.bk_biz_id !== -1 || vpcRelateSubnetCount.value > 0;
});
const bkTooltipsOptions = computed(() => {
  if (isResourcePage.value && detail.value?.bk_biz_id !== -1)
    return {
      content: '该VPC已分配到业务，仅可在业务下操作',
      disabled: detail.value.bk_biz_id === -1,
    };
  if (vpcRelateSubnetCount.value > 0)
    return {
      content: `该vpc关联了 ${vpcRelateSubnetCount.value} 个子网，不可直接删除`,
      disabled: !(vpcRelateSubnetCount.value > 0),
    };

  return null;
});

const handleDeleteVpc = (data: any) => {
  /* const vpcIds = [data.id];
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
      if (cvmsResult?.data?.count || subnetsResult?.data?.count
        || routeResult?.data?.count || networkResult?.data?.count) {
        const getMessage = (result: any, name: string) => {
          if (result?.data?.count) {
            return `${result?.data?.count}个${name}，`;
          }
          return '';
        };
        Message({
          theme: 'error',
          message: `该VPC（name：${data.name}，id：${data.id}）关联${getMessage(cvmsResult, 'CVM')}
            ${getMessage(subnetsResult, '子网')}${getMessage(routeResult, '路由表')}
            ${getMessage(networkResult, '网络接口')}不能删除`,
        });
      } else {*/
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    extCls: 'delete-resource-infobox',
    onConfirm() {
      resourceStore.delete('vpcs', data.id).then(() => {
        Message({
          theme: 'success',
          message: '删除成功',
        });
        router.back();
      });
    },
  });
  //   }
  // });
};
</script>

<template>
  <Teleport to="#breadcrumbExtra">
    <hcm-auth :sign="deleteSign" tag="span" v-slot="{ noPerm }">
      <bk-button
        theme="primary"
        @click="handleDeleteVpc(detail)"
        :disabled="disabledOption || noPerm"
        v-bk-tooltips="bkTooltipsOptions || { disabled: true }"
      >
        {{ t('删除') }}
      </bk-button>
    </hcm-auth>
  </Teleport>

  <bk-loading :loading="loading">
    <div class="detail-content-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
      <detail-tab :tabs="hostTabs">
        <template #default>
          <detail-info
            :detail="detail"
            :fields="VPCFields"
            :label-width="VendorEnum.GCP === detail.vendor ? '200px' : '120px'"
            global-copyable
          />
        </template>
      </detail-tab>

      <detail-tab class="mt16" :tabs="VPCTabs">
        <template #default="type">
          <VPCCidr v-if="type === 'cidr'" :detail="detail" />
          <VPCSubnet v-if="type === 'subnet'" :detail="detail" />
          <VPCRoute v-if="type === 'route'" :detail="detail" />
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

:deep(.detail-info-main .info-list-item .item-field) {
  width: 180px !important;
}
</style>
