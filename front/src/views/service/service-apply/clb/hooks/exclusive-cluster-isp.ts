import type { ExclusiveClusterIsp } from '@/api/load_balancers/apply-clb/types';

/**
 * 独占集群标签接口的 ISP 映射：
 * BGP 与限速 CAP 统一按 BGP 查询，三网直连保留原值。
 */
const EXCLUSIVE_CLUSTER_ISP_MAP: Partial<Record<string, ExclusiveClusterIsp>> = {
  BGP: 'BGP',
  CAP: 'BGP',
  CMCC: 'CMCC',
  CUCC: 'CUCC',
  CTCC: 'CTCC',
};

export const getExclusiveClusterQueryIsp = (isp?: string): ExclusiveClusterIsp | undefined =>
  isp ? EXCLUSIVE_CLUSTER_ISP_MAP[isp] : undefined;

export const isExclusiveClusterIsp = (isp?: string): boolean => Boolean(getExclusiveClusterQueryIsp(isp));
