import { LOAD_BALANCER_ISP_NAME } from '@/views/load-balancer/constants';
import { ExclusiveClusterType } from '@/store/load-balancer/exclusive-cluster';

export const EXCLUSIVE_CLUSTER_TYPE_NAME: Record<ExclusiveClusterType, string> = {
  [ExclusiveClusterType.TGW]: 'TGW',
  [ExclusiveClusterType.STGW]: 'STGW',
};

// 混合运营商：只有独占集群存在该运营商类型（后端 ClusterIsp.MIX）
export const EXCLUSIVE_CLUSTER_MIX_ISP = 'MIX';

// 独占集群运营商比通用运营商多一项「混合运营商」，固定放在末尾。
// 不能把 MIX 加进共享的 LOAD_BALANCER_ISP_NAME，否则负载均衡（CLB）列表的运营商筛选项也会多出该项。
export const EXCLUSIVE_CLUSTER_ISP_NAME = {
  ...LOAD_BALANCER_ISP_NAME,
  [EXCLUSIVE_CLUSTER_MIX_ISP]: '混合运营商',
};

export const EMPTY_DISPLAY = '--';

// 未分配业务
export const UNASSIGNED_BIZ_ID = -1;
