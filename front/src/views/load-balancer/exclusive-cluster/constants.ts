import { LOAD_BALANCER_ISP_NAME } from '@/views/load-balancer/constants';
import { ExclusiveClusterType } from '@/store/load-balancer/exclusive-cluster';

export const EXCLUSIVE_CLUSTER_TYPE_NAME: Record<ExclusiveClusterType, string> = {
  [ExclusiveClusterType.TGW]: 'TGW',
  [ExclusiveClusterType.STGW]: 'STGW',
};

export const EXCLUSIVE_CLUSTER_ISP_NAME = LOAD_BALANCER_ISP_NAME;

export const EMPTY_DISPLAY = '--';

// 未分配业务
export const UNASSIGNED_BIZ_ID = -1;
