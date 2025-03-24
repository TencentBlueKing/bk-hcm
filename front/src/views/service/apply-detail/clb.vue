<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { APPLICATION_TYPE_MAP } from '../apply-list/constants';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { LB_ISP, NET_CHARGE_MAP, VendorMap } from '@/common/constant';
import { LB_NETWORK_TYPE_MAP } from '@/constants';
import { IApplicationDetail } from './index';

import panel from '@/components/panel';
import detailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import gridContainer from '@/components/layout/grid-container/grid-container.vue';
import gridItem from '@/components/layout/grid-container/grid-item.vue';
import status from './components/status.vue';

const DISPLAY_CLB_SPECS_MAP: Record<string, string> = {
  shared: '共享型',
  'clb.c1.small': '性能容量型(clb.c1.small)',
  'clb.c2.medium': '性能容量型(clb.c2.medium)',
  'clb.c3.small': '性能容量型(clb.c3.small)',
  'clb.c3.medium': '性能容量型(clb.c3.medium)',
  'clb.c4.small': '性能容量型(clb.c4.small)',
  'clb.c4.medium': '性能容量型(clb.c4.medium)',
  'clb.c4.large': '性能容量型(clb.c4.large)',
  'clb.c4.xlarge': '性能容量型(clb.c4.xlarge)',
};

const props = defineProps<{ applicationDetail: IApplicationDetail; loading: boolean }>();

const { getNameFromBusinessMap } = useBusinessMapStore();

const clbDetail = computed(() => {
  try {
    const detail = JSON.parse(props.applicationDetail?.content);
    const { zones, backup_zones } = detail;
    Object.assign(detail, {
      zone: backup_zones.length > 0 ? `主备可用区 主(${zones[0]})备(${backup_zones[0]})` : zones.join(','),
    });
    return detail;
  } catch (error) {
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

const paramInfoFields: ModelProperty[] = [
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
  { id: 'sla_type', name: '负载均衡规格类型', type: 'enum', option: DISPLAY_CLB_SPECS_MAP },
  { id: 'internet_charge_type', name: '网络计费模式', type: 'enum', option: NET_CHARGE_MAP },
  { id: 'require_count', name: '需求数量', type: 'number' },
  { id: 'name', name: '实例名称', type: 'string' },
];
</script>

<template>
  <bk-loading v-if="loading" loading style="width: 100%; height: 100%"><div></div></bk-loading>
  <div v-else>
    <detail-header><span>负载均衡申请单详情</span></detail-header>
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
