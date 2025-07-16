import { ref } from 'vue';
import { defineStore } from 'pinia';
import { resolveApiPathByBusinessId } from '@/common/util';
import http from '@/http';

export interface ITargetGroupDetails {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  account_id: string;
  bk_biz_id: number;
  target_group_type: string;
  vpc_id: string;
  cloud_vpc_id: string;
  region: string;
  protocol: string;
  port: number;
  weight: number;
  health_check: {
    health_switch: number;
    time_out: number;
    interval_time: number;
    health_num: number;
    un_health_num: number;
    http_code: number;
    check_type: string;
    source_ip_type: number;
  };
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  target_list: {
    id: string;
    account_id: string;
    ip: string;
    port: number;
    weight: number;
    inst_type: string;
    inst_id: string;
    cloud_inst_id: string;
    inst_name: string;
    target_group_region: string;
    target_group_id: string;
    cloud_target_group_id: string;
    private_ip_address: string[];
    public_ip_address: string[];
    cloud_vpc_ids: string[];
    zone: string;
    memo: string;
    creator: string;
    reviser: string;
    created_at: string;
    updated_at: string;
  }[];
}

export interface ITargetsWeightStatItem {
  target_group_id: string;
  rs_weight_zero_num: number;
  rs_weight_non_zero_num: number;
}

export const useLoadBalancerTargetGroupStore = defineStore('load-balancer-target-group', () => {
  const targetGroupDetailsLoading = ref(false);
  const getTargetGroupDetails = async (id: string, businessId?: number) => {
    targetGroupDetailsLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `target_groups/${id}`, businessId);
    try {
      const res = await http.get(api);
      return res.data as ITargetGroupDetails;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      targetGroupDetailsLoading.value = false;
    }
  };

  const targetGroupRsWeightStatLoading = ref(false);
  const getTargetsWeightStat = async (targetGroupIds: string[], businessId?: number) => {
    targetGroupRsWeightStatLoading.value = true;
    const api = resolveApiPathByBusinessId('/api/v1/cloud', `target_groups/targets/weight_stat`, businessId);
    try {
      const res = await http.post(api, { target_group_ids: targetGroupIds });

      return res.data as ITargetsWeightStatItem[];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      targetGroupRsWeightStatLoading.value = false;
    }
  };

  return {
    targetGroupDetailsLoading,
    getTargetGroupDetails,
    targetGroupRsWeightStatLoading,
    getTargetsWeightStat,
  };
});
