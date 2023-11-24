/**
 * 资源选型 TS 类型
 */

/**
 * 查询业务类型
 */
interface BizType {
  id: string;
  biz_type: string;
  network_latency_tolerance: number;
  deployment_architecture: string[];
}
export interface BizTypeResData {
  code: number;
  message: string;
  data: { details: Array<BizType>; count: number };
}

/**
 * 云资源选型方案
 */
interface AreaInfo {
  name: string;
  value?: number;
  children?: Array<AreaInfo>;
}
export interface GenerateSchemesReqParams {
  cover_ping: number;
  deployment_architecture: Array<'distributed' | 'centralized'>;
  biz_type: string;
  user_distribution: Array<AreaInfo>;
}
interface RecommendScheme {
  cover_rate: number;
  composite_score: number;
  net_score: number;
  cost_score: number;
  result_idc_ids: string[];
}
export interface GenerateSchemesResData {
  code: number;
  message: string;
  data: { details: Array<RecommendScheme>; count: number };
}
