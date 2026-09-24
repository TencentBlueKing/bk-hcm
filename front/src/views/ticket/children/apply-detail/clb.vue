<script setup lang="ts">
import { computed } from 'vue';
import { RouteLocationRaw } from 'vue-router';
import { ModelProperty } from '@/model/typings';
import { APPLICATION_TYPE_MAP } from '@/views/ticket/constants';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { LB_ISP, NET_CHARGE_MAP, VendorMap } from '@/common/constant';
import { LB_NETWORK_TYPE_MAP } from '@/constants';
import { IApplicationDetail } from './index';

import panel from '@/components/panel';
import detailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import gridContainer from '@/components/layout/grid-container/grid-container.vue';
import gridItem from '@/components/layout/grid-container/grid-item.vue';
import status from './components/status.vue';
import { MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { getLoadBalancerInstanceSpecName } from '@/views/load-balancer/utils';

const props = defineProps<{ applicationDetail: IApplicationDetail; loading: boolean }>();

const { getNameFromBusinessMap } = useBusinessMapStore();

const clbDetail = computed(() => {
  try {
    const detail = JSON.parse(props.applicationDetail?.content);
    const { zones, backup_zones } = detail;
    if (backup_zones) {
      Object.assign(detail, {
        zone: backup_zones.length > 0 ? `主备可用区 主(${zones[0]})备(${backup_zones[0]})` : zones.join(','),
      });
    } else {
      Object.assign(detail, { zone: zones?.join(',') });
    }

    const clusters = Array.isArray(detail.clusters) ? detail.clusters : [];
    const l4Cluster = clusters.find(({ cluster_type }: { cluster_type: string }) => cluster_type === 'TGW');
    const l7Cluster = clusters.find(({ cluster_type }: { cluster_type: string }) => cluster_type === 'STGW');
    return {
      ...detail,
      instance_spec: getLoadBalancerInstanceSpecName({
        exclusive: detail.exclusive ?? 0,
        sla_type: detail.sla_type === 'shared' ? '' : detail.sla_type,
      }),
      l4_cluster_tag: l4Cluster?.cluster_tag,
      l4_cluster_name: l4Cluster?.cluster_name,
      l4_cluster_ip: detail.vip,
      l7_cluster_tag: l7Cluster?.cluster_tag || detail.cluster_tag,
    };
  } catch (error) {
    console.error(error);
    return {};
  }
});

const baseInfoFields: ModelProperty[] = [
  { id: 'type', name: '申请类型', type: 'enum', option: APPLICATION_TYPE_MAP },
  { id: 'creator', name: '申请人', type: 'user' },
  { id: 'memo', name: '申请单备注', type: 'string' },
  { id: 'created_at', name: '申请时间', type: 'datetime' },
  { id: 'updated_at', name: '更新时间', type: 'datetime' },
];

const commonParamInfoFields: ModelProperty[] = [
  { id: 'vendor', name: '云厂商', type: 'enum', option: VendorMap },
  { id: 'account_id', name: '云账号', type: 'string' },
  { id: 'load_balancer_type', name: '网络类型', type: 'enum', option: LB_NETWORK_TYPE_MAP },
  { id: 'address_ip_version', name: 'IP版本', type: 'string' },
  { id: 'cloud_vpc_id', name: 'VPC', type: 'string' },
  { id: 'zone', name: '可用区', type: 'string' },
  { id: 'cloud_subnet_id', name: '子网', type: 'string' },
  {
    id: 'zhi_tong',
    name: '直通',
    type: 'bool',
    option: { trueText: '已开启', falseText: '未开启' },
  },
  { id: 'vip_isp', name: '运营商类型', type: 'enum', option: LB_ISP },
  { id: 'tgw_group_name', name: '免流', type: 'string' },
  { id: 'instance_spec', name: '负载均衡规格类型', type: 'string' },
];

const exclusiveClusterFields: ModelProperty[] = [
  { id: 'l4_cluster_tag', name: '四层集群标签', type: 'string' },
  { id: 'l4_cluster_name', name: '四层集群名称', type: 'string' },
  { id: 'l4_cluster_ip', name: '四层集群 IP', type: 'string' },
  { id: 'l7_cluster_tag', name: '七层集群标签', type: 'string' },
];

const remainingParamInfoFields: ModelProperty[] = [
  { id: 'internet_charge_type', name: '网络计费模式', type: 'enum', option: NET_CHARGE_MAP },
  { id: 'require_count', name: '需求数量', type: 'number' },
  { id: 'name', name: '实例名称', type: 'string' },
];

const paramInfoFields = computed(() => [
  ...commonParamInfoFields,
  ...(clbDetail.value.exclusive === 1 ? exclusiveClusterFields : []),
  ...remainingParamInfoFields,
]);

const navigateTo: RouteLocationRaw = {
  name: MENU_SERVICE_TICKET_MANAGEMENT,
  query: { type: 'load-balancer' },
};
</script>

<template>
  <bk-loading v-if="loading" loading style="width: 100%; height: 100%"><div></div></bk-loading>
  <div v-else>
    <detail-header :to="navigateTo"><span>负载均衡申请单详情</span></detail-header>
    <div class="container">
      <status :application-detail="applicationDetail" />
      <panel title="基本信息">
        <grid-container fixed :column="2" :content-min-width="300" :label-width="150">
          <grid-item label="业务名称">
            {{ clbDetail?.bk_biz_id !== -1 ? getNameFromBusinessMap(clbDetail.bk_biz_id) : '未分配' }}
          </grid-item>
          <grid-item v-for="field in baseInfoFields" :key="field.id" :label="field.name">
            <display-value
              :property="field"
              :value="applicationDetail[field.id]"
              :display="{ ...field.meta?.display, on: 'info' }"
            />
          </grid-item>
        </grid-container>
      </panel>
      <panel title="参数信息">
        <grid-container fixed :column="2" :content-min-width="300" :label-width="150">
          <grid-item v-for="field in paramInfoFields" :key="field.id" :label="field.name">
            <display-value
              :property="field"
              :value="clbDetail[field.id]"
              :display="{ ...field.meta?.display, on: 'info' }"
            />
          </grid-item>
        </grid-container>
      </panel>
    </div>
  </div>
</template>

<style scoped lang="scss">
.container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 52px;
  background-color: #f5f7fa;
  padding: 24px;
}
</style>
